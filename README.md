<div align="center">

<img src="https://raw.githubusercontent.com/iSreyanshu/tgrammY/app/assets/burn-address.png" alt="wallet-generator" style="width: 70%; border-radius: 10px;" />

# Wallet Generator

**Ethereum wallet generator written in Go**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-purple?style=for-the-badge)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-blue?style=for-the-badge)]()

</div>

---

## <img src="https://raw.githubusercontent.com/iSreyanshu/tgrammY/app/assets/turtle-banner.png" width="50" align="center" /> Features

| Feature | Description |
|---|---|
| **Multi-core generation** | 20 goroutines hammering ECDSA key-pairs in parallel |
| **Vanity addresses** | Custom prefix search up to 10 hex characters |
| **No-F filter** | Automatically skips any address containing `f` or `F` |
| **Auto-save** | Every wallet is appended to `wallets.txt` instantly |
| **Graceful shutdown** | `Ctrl+C` cleanly stops all workers mid-run |

---

## <img src="https://raw.githubusercontent.com/iSreyanshu/tgrammY/app/assets/turtle-banner.png" width="50" align="center" /> Preview

```shell
  Select mode:
  [1]  Normal wallets   - fast bulk generation, no-F filter
  [2]  Vanity wallets   - custom prefix + no-F filter

  ❯ 1

  Address   : 0x1a2b3c4d5e6a7b8c9d0e1a2b3c4d5e6a7b8c9d0e
  PrivateKey: 4f3e2d1c0b9a8f7e6d5c4b3a2918273645...
  Generated : 14:32:01

  ✔ Done! Generated 100 wallets in 1.23s (81 wallets/sec)
  💾 Saved to wallets.txt
```

---

## <img src="https://raw.githubusercontent.com/iSreyanshu/tgrammY/app/assets/turtle-banner.png" width="50" align="center" /> Quick Start

### 1. Clone

```shell
git clone https://github.com/iSreyanshu/Wallets.git
cd Wallets
```

### 2. Install dependencies

```shell
go mod init walletgen
go get github.com/ethereum/go-ethereum
go get github.com/schollz/progressbar/v3
```

### 3. Run

```shell
go run gen.go
```

---

## <img src="https://raw.githubusercontent.com/iSreyanshu/tgrammY/app/assets/turtle-banner.png" width="50" align="center" /> Modes

### Mode1 - Normal Wallets
Generates N wallets as fast as possible. All addresses are filtered to exclude any containing the letter `F`.

### Mode2 - Vanity Wallets
Searches for addresses matching your custom prefix. Example: entering `dead` will find addresses starting with `0xdead…`

> **Tip:** Longer prefixes = exponentially more attempts. Keep it under 6 chars for reasonable speed.

---

## <img src="https://raw.githubusercontent.com/iSreyanshu/tgrammY/app/assets/turtle-banner.png" width="50" align="center" /> Output

All wallets are saved to **`wallets.txt`** in the same directory:

```shell
Address: 0x1a2b... | PrivateKey: 4f3e... | Generated: 2026-05-02T14:32:01Z
```

---

## <img src="https://raw.githubusercontent.com/iSreyanshu/tgrammY/app/assets/turtle-banner.png" width="50" align="center" /> Security Notice

> **Never share your private keys.**  
> This tool is for educational and development purposes.  
> Store `wallets.txt` securely and delete it when no longer needed.

---

## <img src="https://raw.githubusercontent.com/iSreyanshu/tgrammY/app/assets/turtle-banner.png" width="50" align="center" /> Configuration

Edit the `Config` struct in `gen.go` to tune performance:

```go
cfg := Config{
    WorkerCount: 20, // Increase for more CPU cores
}
```

---

<div align="center">
  
  Made with ❤️ by [iSreyanshu](https://github.com/iSreyanshu)
  
</div>
