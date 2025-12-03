package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/schollz/progressbar/v3"
)

type Config struct {
	WorkerCount int
}

type Wallet struct {
	Address    string
	PrivateKey string
	CreatedAt  time.Time
}

func main() {
	config := Config{
		WorkerCount: 20, // CURRENTLY USING 20 GOROUTINS
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
		generateNormalWallets(config, count)

	case 2:
		prefix, count, err := getVanityParams()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\033[1;31mError reading input: %v\033[0m\n", err)
			os.Exit(1)
		}

		if len(prefix) > 10 {
			fmt.Println("\033[1;31mPrefix too long! Max 10 hex characters allowed.\033[0m")
			os.Exit(1)
		}

		generateVanityWallets(config, prefix, count)

	default:
		fmt.Println("\033[1;31mInvalid choice.\033[0m")
		os.Exit(1)
	}
}

func getUserChoice() (int, error) {
	fmt.Println("\n\033[1;36mChoose an option:\033[0m")
	fmt.Println("1. Generate normal wallets (no F allowed)")
	fmt.Println("2. Generate vanity wallets (prefix + no F allowed)")
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
	fmt.Print("\033[1;36mEnter vanity prefix (max 10 chars after 0x): >> 0x\033[0m")
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

func isValidAddress(addr string) bool {
	return !strings.Contains(addr, "f") && !strings.Contains(addr, "F")
}

func generateNormalWallets(config Config, count int) {
	results := make(chan Wallet)

	bar := progressbar.NewOptions(count,
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetDescription("Generating wallets:"),
		progressbar.OptionShowCount(),
	)

	for i := 0; i < config.WorkerCount; i++ {
		go func() {
			for {
				key, _ := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
				addr := strings.ToLower(crypto.PubkeyToAddress(key.PublicKey).Hex())

				if !isValidAddress(addr) {
					continue
				}

				results <- Wallet{
					Address:    addr,
					PrivateKey: fmt.Sprintf("%x", crypto.FromECDSA(key)),
					CreatedAt:  time.Now(),
				}
			}
		}()
	}

	for i := 0; i < count; i++ {
		w := <-results
		fmt.Printf("\nAddress: %s\nPrivateKey: %s\n\n", w.Address, w.PrivateKey)
		bar.Add(1)
	}
}

func generateVanityWallets(config Config, prefix string, count int) {
	results := make(chan Wallet)
	done := make(chan struct{})

	fullPrefix := strings.ToLower("0x" + prefix)

	bar := progressbar.NewOptions(count,
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetDescription("Generating vanity wallets:"),
		progressbar.OptionShowCount(),
	)

	for i := 0; i < config.WorkerCount; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					key, _ := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
					addr := strings.ToLower(crypto.PubkeyToAddress(key.PublicKey).Hex())

					if !strings.HasPrefix(addr, fullPrefix) {
						continue
					}
					if !isValidAddress(addr) {
						continue
					}

					results <- Wallet{
						Address:    addr,
						PrivateKey: fmt.Sprintf("%x", crypto.FromECDSA(key)),
						CreatedAt:  time.Now(),
					}
				}
			}
		}()
	}

	for i := 0; i < count; i++ {
		w := <-results
		fmt.Printf("\nAddress: %s\nPrivateKey: %s\n\n", w.Address, w.PrivateKey)
		bar.Add(1)
	}

	close(done)
}
