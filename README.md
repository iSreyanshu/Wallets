# evm

![Zig](https://camo.githubusercontent.com/d76765c2b9b969ac751fd3327d8e5105f5ad248438194717fce38c025a90e5d8/68747470733a2f2f696d672e736869656c64732e696f2f62616467652f5a69672d302e31332e302d6637613431643f6c6f676f3d7a6967266c6f676f436f6c6f723d7768697465)
![Crypto](https://camo.githubusercontent.com/b78f744e5f520e134d15e1ea7176c8bc5c666bafbdb52f5981e54d8f61f7c448/68747470733a2f2f696d672e736869656c64732e696f2f62616467652f43727970746f2d736563703235366b312532302532422532304b656363616b2d326561303433)

**A local evm wallet generator with normal, vanity and exclusion modes.**
It uses operating-system randomness, OpenSSL secp256k1, Ethereum Keccak-256 and a reusable Zig thread pool.

## Quick Start

```sh
git clone https://github.com/iSreyanshu/Wallets.git
cd Wallets
chmod +x install.sh run.sh
./install.sh
./run.sh
```

The installer handles compiler, build and OpenSSL prerequisites. It installs
Zig in `~/.local/zig` and does not require a system-wide Zig package.

## Modes

Each run opens an interactive menu and asks for the wallet count:

1. Generate normal wallets.
2. Generate vanity wallets by matching up to 10 hexadecimal characters after
	`0x` or at the end of the address.
3. Generate wallets excluding a comma-separated list of hexadecimal characters,
	such as `b,6,f`.

The output is written to `wallets.csv` (address has no `0x`; private key keeps
its `0x` prefix):

```text
address,private_key
023bcd...,0x...
```

## Performance

The default pool size is **100 workers per detected CPU core** capped by the
requested wallet count. Override it with `--worker`:

```sh
./run.sh --count 100000 --worker 1000
```

## Manual Build

The build links OpenSSL's secp256k1 implementation and includes the Ethereum
Keccak-256 implementation in `deps/keccak`.

```sh
zig build -Doptimize=ReleaseFast
zig build run -Doptimize=ReleaseFast -- --count 1000 --worker 400
```

## Security

- Private keys come from `std.crypto.random`, backed by the operating system.
- `wallets.csv` contains spendable secrets. Treat it like a password vault.
- Never commit, upload or paste generated private keys into chat or issue trackers.
- This tool does not connect to a blockchain or check balances.
