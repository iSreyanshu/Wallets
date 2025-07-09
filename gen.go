package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/schollz/progressbar/v3"
)

type Config struct {
	WorkerCount int
	OutputDir   string
}

type Wallet struct {
	Address    string
	PrivateKey string
	CreatedAt  time.Time
}

func main() {
	config := Config{
		WorkerCount: 15,
		OutputDir:   "wallets",
	}

	fmt.Println("\033[1;34m=== Ethereum Wallet Generator ===\033[0m")
	choice, err := getUserChoice()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[1;31mError reading choice: %v\033[0m\n", err)
		os.Exit(1)
	}

	switch choice {
	case 1:
		count, err := getWalletCount()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\033[1;31mError reading count: %v\033[0m\n", err)
			os.Exit(1)
		}
		if count <= 0 {
			fmt.Println("\033[1;31mCount must be positive.\033[0m")
			os.Exit(1)
		}
		if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "\033[1;31mError creating output directory: %v\033[0m\n", err)
			os.Exit(1)
		}
		startTime := time.Now()
		fmt.Printf("\033[1;33mEstimated Time: ~%d seconds\033[0m\n", count/1000)
		generateNormalWallets(config, count)
		fmt.Printf("\033[1;32mTime Taken: %v\033[0m\n", time.Since(startTime))
	case 2:
		prefix, count, err := getVanityParams()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\033[1;31mError reading input: %v\033[0m\n", err)
			os.Exit(1)
		}
		if len(prefix) > 6 {
			fmt.Println("\033[1;31mPrefix too long. Max 6 characters allowed.\033[0m")
			os.Exit(1)
		}
		if count <= 0 {
			fmt.Println("\033[1;31mCount must be positive.\033[0m")
			os.Exit(1)
		}
		if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "\033[1;31mError creating output directory: %v\033[0m\n", err)
			os.Exit(1)
		}
		estimate := float64(count*(1<<(len(prefix)*4))/1000)
		fmt.Printf("\033[1;33mEstimated Time: ~%.0f seconds\033[0m\n", estimate)
		generateVanityWallets(config, prefix, count)
	default:
		fmt.Println("\033[1;31mInvalid choice.\033[0m")
		os.Exit(1)
	}
}

func getUserChoice() (int, error) {
	fmt.Println("\n\033[1;36mChoose an option:\033[0m")
	fmt.Println("1. Generate default wallets")
	fmt.Println("2. Generate vanity wallet addresses")
	fmt.Print("\033[1;32m>> \033[0m")
	var choice int
	_, err := fmt.Scan(&choice)
	return choice, err
}

func getWalletCount() (int, error) {
	fmt.Print("\033[1;36mHow many wallets to generate? >> \033[0m")
	var count int
	_, err := fmt.Scan(&count)
	return count, err
}

func getVanityParams() (string, int, error) {
	fmt.Print("\033[1;36mEnter vanity prefix (max 6 hex chars after 0x): >> 0x\033[0m")
	var prefix string
	_, err := fmt.Scan(&prefix)
	if err != nil {
		return "", 0, err
	}
	fmt.Print("\033[1;36mHow many vanity wallets to generate? >> \033[0m")
	var count int
	_, err = fmt.Scan(&count)
	return prefix, count, err
}

func generateNormalWallets(config Config, count int) {
	var wg sync.WaitGroup
	jobs := make(chan int, count)
	results := make(chan Wallet, count)
	outputFile, err := os.Create(filepath.Join(config.OutputDir, fmt.Sprintf("wallets_%s.txt", time.Now().Format("20060102_150405"))))
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[1;31mError creating output file: %v\033[0m\n", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	bar := progressbar.NewOptions(count,
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetDescription("\033[1;34mGenerating wallets:\033[0m"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	workerCount := config.WorkerCount
	if count < workerCount {
		workerCount = count
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for range jobs {
				key, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
				if err != nil {
					fmt.Fprintf(os.Stderr, "\033[1;31mWorker %d: Error generating key: %v\033[0m\n", workerID, err)
					continue
				}
				wallet := Wallet{
					Address:    crypto.PubkeyToAddress(key.PublicKey).Hex(),
					PrivateKey: fmt.Sprintf("%x", crypto.FromECDSA(key)),
					CreatedAt:  time.Now(),
				}
				results <- wallet
			}
		}(i)
	}

	go func() {
		for i := 0; i < count; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	go func() {
		defer close(results)
		for i := 0; i < count; i++ {
			wallet := <-results
			output := fmt.Sprintf("address: %s\nPrivate key: %s\nCreated at: %s\nProgress: %d/%d\n\n",
				wallet.Address, wallet.PrivateKey, wallet.CreatedAt.Format(time.RFC3339), i+1, count)
			fmt.Print(output)
			fmt.Fprint(outputFile, output)
			bar.Add(1)
		}
	}()

	wg.Wait()
}

func generateVanityWallets(config Config, prefix string, count int) {
	var wg sync.WaitGroup
	results := make(chan Wallet, count)
	done := make(chan struct{})
	prefix = strings.ToLower("0x" + prefix)
	outputFile, err := os.Create(filepath.Join(config.OutputDir, fmt.Sprintf("vanity_wallets_%s.txt", time.Now().Format("20060102_150405"))))
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[1;31mError creating output file: %v\033[0m\n", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	bar := progressbar.NewOptions(count,
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetDescription("\033[1;34mGenerating vanity wallets:\033[0m"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	for i := 0; i < config.WorkerCount; i++ {
		go func(workerID int) {
			for {
				select {
				case <-done:
					return
				case <-time.After(time.Hour):
					fmt.Fprintf(os.Stderr, "\033[1;31mWorker %d: Timeout reached\033[0m\n", workerID)
					return
				default:
					key, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
					if err != nil {
						fmt.Fprintf(os.Stderr, "\033[1;31mWorker %d: Error generating key: %v\033[0m\n", workerID, err)
						continue
					}
					address := strings.ToLower(crypto.PubkeyToAddress(key.PublicKey).Hex())
					if strings.HasPrefix(address, prefix) {
						select {
						case results <- Wallet{
							Address:    address,
							PrivateKey: fmt.Sprintf("%x", crypto.FromECDSA(key)),
							CreatedAt:  time.Now(),
						}:
						default:
						}
					}
				}
			}
		}(i)
	}

	go func() {
		for i := 0; i < count; i++ {
			wg.Add(1)
			wallet := <-results
			output := fmt.Sprintf("address: %s\nPrivate key: %s\nCreated at: %s\nProgress: %d/%d\n\n",
				wallet.Address, wallet.PrivateKey, wallet.CreatedAt.Format(time.RFC3339), i+1, count)
			fmt.Print(output)
			fmt.Fprint(outputFile, output)
			bar.Add(1)
			wg.Done()
		}
		close(done)
		close(results)
	}()

	wg.Wait()
}
