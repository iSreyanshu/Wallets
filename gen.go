package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/schollz/progressbar/v3"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	red    = "\033[1;31m"
	green  = "\033[1;32m"
	yellow = "\033[1;33m"
	cyan   = "\033[1;36m"
	blue   = "\033[1;34m"
	purple = "\033[1;35m"
	dim    = "\033[2m"
)

func colorize(color, text string) string { return color + text + reset }

type Config struct {
	WorkerCount int
}

type Wallet struct {
	Address    string
	PrivateKey string
	CreatedAt  time.Time
}

func (w Wallet) String() string {
	ts := w.CreatedAt.Format("15:04:05")
	return fmt.Sprintf(
		"\n  %s %s\n  %s %s\n  %s %s\n",
		colorize(cyan, "Address   :"), colorize(green, w.Address),
		colorize(yellow, "PrivateKey:"), w.PrivateKey,
		colorize(dim, "Generated :"), ts,
	)
}

func main() {
	cfg := Config{WorkerCount: 20}

	choice, err := promptChoice()
	if err != nil {
		fatal("reading choice", err)
	}

	switch choice {
	case 1:
		count, err := promptCount("How many wallets to generate?")
		if err != nil {
			fatal("reading count", err)
		}
		generateNormal(cfg, count)

	case 2:
		prefix, count, err := promptVanity()
		if err != nil {
			fatal("reading vanity params", err)
		}
		if len(prefix) > 10 {
			fmt.Println(colorize(red, "  ✗ Prefix too long - max 10 hex characters."))
			os.Exit(1)
		}
		generateVanity(cfg, prefix, count)

	default:
		fmt.Println(colorize(red, "  ✗ Invalid choice."))
		os.Exit(1)
	}
}


func promptChoice() (int, error) {
	fmt.Println(colorize(cyan, "  Select mode:"))
	fmt.Println(colorize(green, "  [1]") + "  Normal wallets   - fast bulk generation, no-F filter")
	fmt.Println(colorize(green, "  [2]") + "  Vanity wallets   - custom prefix + no-F filter")
	fmt.Print(colorize(bold, "\n  ❯ "))
	var c int
	_, err := fmt.Scan(&c)
	return c, err
}

func promptCount(label string) (int, error) {
	fmt.Printf("\n  %s %s", colorize(cyan, label), colorize(bold, "❯ "))
	var n int
	_, err := fmt.Scan(&n)
	return n, err
}

func promptVanity() (string, int, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\n  %s %s", colorize(cyan, "Vanity prefix (max 10 hex chars after 0x):"), colorize(bold, "0x"))
	prefix, err := reader.ReadString('\n')
	if err != nil {
		return "", 0, err
	}
	prefix = strings.TrimSpace(prefix)

	count, err := promptCount("How many vanity wallets?")
	return prefix, count, err
}

// isClean returns true when the address contains no 'f' character.
func isClean(addr string) bool {
	return !strings.ContainsAny(addr, "fF")
}

// generateKey creates a fresh ECDSA key and returns the lowercase address + hex private key.
func generateKey() (addr, privHex string, err error) {
	key, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return "", "", err
	}
	addr = strings.ToLower(crypto.PubkeyToAddress(key.PublicKey).Hex())
	privHex = fmt.Sprintf("%x", crypto.FromECDSA(key))
	return addr, privHex, nil
}

// saveWallet appends a wallet to wallets.txt.
func saveWallet(w Wallet) {
	f, err := os.OpenFile("wallets.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "Address: %s | PrivateKey: %s | Generated: %s\n",
		w.Address, w.PrivateKey, w.CreatedAt.Format(time.RFC3339))
}

func generateNormal(cfg Config, count int) {
	fmt.Printf("\n  %s Spinning up %s workers…\n\n",
		colorize(yellow, "⚙"), colorize(bold, fmt.Sprintf("%d", cfg.WorkerCount)))

	results := make(chan Wallet, cfg.WorkerCount*2)
	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Graceful Ctrl-C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println(colorize(yellow, "\n\n  ⚠  Interrupted - closing workers…"))
		close(stop)
	}()

	for i := 0; i < cfg.WorkerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					addr, priv, err := generateKey()
					if err != nil || !isClean(addr) {
						continue
					}
					results <- Wallet{Address: addr, PrivateKey: priv, CreatedAt: time.Now()}
				}
			}
		}()
	}

	go func() { wg.Wait(); close(results) }()

	bar := newBar(count, "Generating wallets")
	start := time.Now()
	collected := 0

	for w := range results {
		if collected >= count {
			break
		}
		fmt.Print(w)
		saveWallet(w)
		bar.Add(1)
		collected++
	}

	close(stop)
	printStats(collected, start)
}

func generateVanity(cfg Config, prefix string, count int) {
	fullPrefix := strings.ToLower("0x" + prefix)

	fmt.Printf("\n  %s Hunting for prefix %s with %s workers…\n\n",
		colorize(yellow, "⚙"),
		colorize(green, fullPrefix),
		colorize(bold, fmt.Sprintf("%d", cfg.WorkerCount)),
	)

	results := make(chan Wallet, cfg.WorkerCount*2)
	var wg sync.WaitGroup
	stop := make(chan struct{})
	var attempts atomic.Int64

	// Graceful Ctrl-C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println(colorize(yellow, "\n\n  ⚠  Interrupted - closing workers…"))
		close(stop)
	}()

	for i := 0; i < cfg.WorkerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					attempts.Add(1)
					addr, priv, err := generateKey()
					if err != nil {
						continue
					}
					if !strings.HasPrefix(addr, fullPrefix) || !isClean(addr) {
						continue
					}
					results <- Wallet{Address: addr, PrivateKey: priv, CreatedAt: time.Now()}
				}
			}
		}()
	}

	go func() { wg.Wait(); close(results) }()

	bar := newBar(count, "Vanity search")
	start := time.Now()
	collected := 0

	for w := range results {
		if collected >= count {
			break
		}
		fmt.Print(w)
		saveWallet(w)
		bar.Add(1)
		collected++
		fmt.Printf("  %s %s attempts so far\n", colorize(dim, "↳"), colorize(yellow, fmt.Sprintf("%d", attempts.Load())))
	}

	close(stop)
	printStats(collected, start)
	fmt.Printf("  %s Total attempts: %s\n",
		colorize(cyan, "🔍"), colorize(bold, fmt.Sprintf("%d", attempts.Load())))
}

func newBar(total int, desc string) *progressbar.ProgressBar {
	return progressbar.NewOptions(total,
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetDescription(colorize(cyan, "  "+desc+":")),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        colorize(green, "█"),
			SaucerPadding: colorize(dim, "░"),
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
	)
}

func printStats(count int, start time.Time) {
	elapsed := time.Since(start)
	rate := float64(count) / elapsed.Seconds()
	fmt.Printf("\n\n  %s Done! Generated %s wallets in %s (%.0f wallets/sec)\n",
		colorize(green, "✔"),
		colorize(bold, fmt.Sprintf("%d", count)),
		colorize(yellow, elapsed.Round(time.Millisecond).String()),
		rate,
	)
	fmt.Printf("  %s Saved to %s\n\n", colorize(cyan, "💾"), colorize(bold, "wallets.txt"))
}

func fatal(context string, err error) {
	fmt.Fprintf(os.Stderr, "%s %s: %v\n", colorize(red, "  ✗ Error"), context, err)
	os.Exit(1)
}
