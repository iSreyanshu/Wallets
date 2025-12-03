# Ethereum Wallet Generator (Fast Vanity and Normal Wallets)

A high-performance Ethereum wallet generator written in Go, supporting:

- Multi-threaded wallet generation  
- Vanity address search (custom prefix)  
- Filtering wallets that contain **F**

---

### Generate normal wallets  
Fast multi-core ECDSA key generation.

### Generate vanity wallets  
Supports custom prefixes up to 10 hex characters.

```sh
git clone https://github.com/iYashKun/Wallets.git
cd Wallets

go version
go mod init walletgen
go get github.com/ethereum/go-ethereum
go get github.com/schollz/progressbar/v3
go run gen.go
```
