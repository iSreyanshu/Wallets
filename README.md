# Ethereum Wallet Generator

A high-performance Ethereum wallet generator written in Go, supporting:

- Multi-threaded wallet generation  
- Vanity address search (custom prefixs)  
- Filtering wallets that contain **F**

---

### Generate normal wallets  
Fast multi-core ECDSA key generation.

### Generate vanity wallets  
Supports custom prefixes up to 10 hex characters.

```sh
git clone https://github.com/iYashKun/Wallets.git
```
```sh
cd Wallets
```
```sh
go version
```
```sh
go mod init walletgen
```
```sh
go get github.com/ethereum/go-ethereum
```
```sh
go get github.com/schollz/progressbar/v3
```
```sh
go run gen.go
```
