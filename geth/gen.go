package main

import (
    "bufio"
    crand "crypto/rand"
    "encoding/hex"
    "flag"
    "fmt"
    "io"
    "os"
    "runtime"
    "strings"
    "sync"
    "sync/atomic"
    "time"

    "github.com/ethereum/go-ethereum/crypto"
)

var totalChecked uint64

func hexToNibbles(hexStr string) []byte {
    hexStr = strings.TrimPrefix(strings.ToLower(hexStr), "0x")
    nibbles := make([]byte, len(hexStr))
    for i, ch := range hexStr {
        switch {
        case ch >= '0' && ch <= '9':
            nibbles[i] = byte(ch - '0')
        case ch >= 'a' && ch <= 'f':
            nibbles[i] = byte(ch - 'a' + 10)
        }
    }
    return nibbles
}

func addrHasNoF(addr [20]byte) bool {
    for _, b := range addr {
        if (b>>4) == 0x0F || (b&0x0F) == 0x0F {
            return false
        }
    }
    return true
}

func addrMatchesPrefix(addr [20]byte, prefixNibbles []byte) bool {
    for i, nib := range prefixNibbles {
        byteIdx := i / 2
        var actual byte
        if i%2 == 0 {
            actual = addr[byteIdx] >> 4
        } else {
            actual = addr[byteIdx] & 0x0F
        }
        if actual != nib {
            return false
        }
    }
    return true
}

func addrMatchesSuffix(addr [20]byte, suffixNibbles []byte) bool {
    startNib := 40 - len(suffixNibbles)
    for i, nib := range suffixNibbles {
        nibIdx := startNib + i
        byteIdx := nibIdx / 2
        var actual byte
        if nibIdx%2 == 0 {
            actual = addr[byteIdx] >> 4
        } else {
            actual = addr[byteIdx] & 0x0F
        }
        if actual != nib {
            return false
        }
    }
    return true
}

func toChecksumAddr(addrHex string) string {
    addr := strings.TrimPrefix(addrHex, "0x")
    if len(addr) < 40 {
        addr = strings.Repeat("0", 40-len(addr)) + addr
    }
    lower := strings.ToLower(addr)
    hash := crypto.Keccak256([]byte(lower))
    hashHex := hex.EncodeToString(hash)

    result := make([]byte, 42)
    result[0] = '0'
    result[1] = 'x'
    for i := 0; i < 40; i++ {
        c := lower[i]
        if c >= 'a' && c <= 'f' {
            nib := hashHex[i]
            if nib >= '8' {
                c = c - 'a' + 'A'
            }
        }
        result[i+2] = c
    }
    return string(result)
}

func formatNumber(n int64) string {
    switch {
    case n >= 1_000_000_000:
        return fmt.Sprintf("%.2fB", float64(n)/1_000_000_000)
    case n >= 1_000_000:
        return fmt.Sprintf("%.2fM", float64(n)/1_000_000)
    case n >= 1_000:
        return fmt.Sprintf("%.2fK", float64(n)/1_000)
    default:
        return fmt.Sprintf("%d", n)
    }
}

func formatDuration(seconds float64) string {
    switch {
    case seconds < 60:
        return fmt.Sprintf("%.0fs", seconds)
    case seconds < 3600:
        return fmt.Sprintf("%.1fm", seconds/60)
    default:
        return fmt.Sprintf("%.1fh", seconds/3600)
    }
}

func printBanner(prefix, suffix string, noF bool, threads int) {
    fmt.Println()
    fmt.Println("  Ethereum Vanity Address Generator:")
    fmt.Printf("  Prefix:       0x%s\n", prefix)
    if suffix != "" {
        fmt.Printf("  Suffix:       %s\n", suffix)
    }
    fmt.Printf("  No 'f':       %v\n", noF)
    fmt.Printf("  Threads:      %d (CPU cores: %d)\n", threads, runtime.NumCPU())
    fmt.Println("------------------------------------------------------------")
}

func statsLoop(expected uint64, stop chan struct{}, startTime time.Time) {
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    lastChecked := uint64(0)
    lastTime := time.Now()

    for {
        select {
        case <-stop:
            return
        case <-ticker.C:
            checked := atomic.LoadUint64(&totalChecked)
            now := time.Now()
            elapsed := now.Sub(startTime).Seconds()
            interval := now.Sub(lastTime).Seconds()

            if elapsed < 0.5 {
                continue
            }

            var speedInst float64
            if interval > 0 {
                speedInst = float64(checked-lastChecked) / interval
            }
            speedAvg := float64(checked) / elapsed

            remaining := expected - (checked % expected)
            if speedAvg > 0 {
                eta := float64(remaining) / speedAvg
                pct := float64(checked%expected) / float64(expected) * 100

                barWidth := 30
                filled := int(pct / 100 * float64(barWidth))
                bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

                fmt.Printf("\r  %s %5.1f%% | %s/s (avg: %s) | ETA: %s     ",
                    bar, pct,
                    formatNumber(int64(speedInst)),
                    formatNumber(int64(speedAvg)),
                    formatDuration(eta))
            }

            lastChecked = checked
            lastTime = now
        }
    }
}

type Result struct {
    PrivKey string
    Addr    [20]byte
}

func worker(chunkSize uint64, prefixNibbles, suffixNibbles []byte, noF bool, wg *sync.WaitGroup, found *uint32, resultChan chan<- Result) {
    defer wg.Done()

    reader := bufio.NewReaderSize(crand.Reader, 1024*1024)
    buf := make([]byte, 32)

    for i := uint64(0); i < chunkSize; i++ {
        if atomic.LoadUint32(found) == 1 {
            return
        }

        _, err := io.ReadFull(reader, buf)
        if err != nil {
            continue
        }

        priv, err := crypto.ToECDSA(buf)
        if err != nil {
            continue
        }

        addr := crypto.PubkeyToAddress(priv.PublicKey)
        match := addrMatchesPrefix(addr, prefixNibbles)
        if !match {
            atomic.AddUint64(&totalChecked, 1)
            continue
        }

        if suffixNibbles != nil {
            match = match && addrMatchesSuffix(addr, suffixNibbles)
            if !match {
                atomic.AddUint64(&totalChecked, 1)
                continue
            }
        }

        if noF {
            match = match && addrHasNoF(addr)
            if !match {
                atomic.AddUint64(&totalChecked, 1)
                continue
            }
        }

        if match && atomic.CompareAndSwapUint32(found, 0, 1) {
            privHex := hex.EncodeToString(crypto.FromECDSA(priv))
            resultChan <- Result{PrivKey: privHex, Addr: addr}
            return
        }

        atomic.AddUint64(&totalChecked, 1)
    }
}

func main() {
    var (
        prefixFlag  = flag.String("prefix", "", "Desired address prefix without 0x, e.g., 0000")
        suffixFlag  = flag.String("suffix", "", "Desired address suffix without 0x (optional)")
        noFFlag     = flag.Bool("no-f", false, "Exclude 'f' anywhere in the address")
        threadsFlag = flag.Int("threads", 0, "Number of goroutines (0 = NumCPU)")
        chunkFlag   = flag.Uint64("chunk", 50000, "Keys per worker batch")
    )
    flag.Parse()

    if *prefixFlag == "" {
        fmt.Println("Usage: go run gen.go -prefix 0000 [-suffix 2222] [-no-f] [-threads 16]")
        os.Exit(1)
    }

    threads := *threadsFlag
    if threads == 0 {
        threads = runtime.NumCPU()
    }
    runtime.GOMAXPROCS(threads)

    prefixNibbles := hexToNibbles(*prefixFlag)
    var suffixNibbles []byte
    if *suffixFlag != "" {
        suffixNibbles = hexToNibbles(*suffixFlag)
    }

    printBanner(*prefixFlag, *suffixFlag, *noFFlag, threads)

    expected := uint64(1)
    for i := 0; i < len(prefixNibbles); i++ {
        expected *= 16
    }
    if suffixNibbles != nil {
        for i := 0; i < len(suffixNibbles); i++ {
            expected *= 16
        }
    }
    if *noFFlag {
        expected = uint64(float64(expected) / 0.076)
    }

    fmt.Printf("  Expected:     ~%s attempts\n", formatNumber(int64(expected)))
    fmt.Println("------------------------------------------------------------")
    fmt.Println()

    startTime := time.Now()
    stopStats := make(chan struct{})
    go statsLoop(expected, stopStats, startTime)

    chunkSize := *chunkFlag
    var found uint32
    resultChan := make(chan Result, 1)

    for found == 0 {
        var wg sync.WaitGroup
        for t := 0; t < threads; t++ {
            wg.Add(1)
            go worker(chunkSize, prefixNibbles, suffixNibbles, *noFFlag, &wg, &found, resultChan)
        }
        wg.Wait()
    }

    close(stopStats)
    elapsed := time.Since(startTime)

    fmt.Println()
    fmt.Println()

    res := <-resultChan
    checked := atomic.LoadUint64(&totalChecked)

    fmt.Println("                       FOUND!")
    fmt.Printf("  Private Key:  0x%s\n", res.PrivKey)
    fmt.Printf("  Address:      %s\n", toChecksumAddr(hex.EncodeToString(res.Addr[:])))
    fmt.Println("------------------------------------------------------------")
    fmt.Printf("  Attempts:     %s\n", formatNumber(int64(checked)))
    fmt.Printf("  Time:         %.2fs\n", elapsed.Seconds())
    if elapsed.Seconds() > 0 {
        fmt.Printf("  Avg Speed:    %s keys/sec\n", formatNumber(int64(float64(checked)/elapsed.Seconds())))
    }
    fmt.Println()
}
