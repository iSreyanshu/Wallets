# EVM wallet generator

Fast, local EVM wallet generation in Zig. It uses secure randomness, OpenSSL
secp256k1, Ethereum Keccak-256, and a reusable thread pool. The default pool
size is 100 workers per detected CPU core, capped by the requested wallet count.

Each run opens an interactive menu:

1. Generate normal wallets.
2. Generate vanity wallets by matching up to 10 hexadecimal characters after
	`0x` or at the end of the address.
3. Generate wallets excluding a comma-separated list of hexadecimal characters,
	such as `b,6,f`.

The output is written to `wallets.csv`:

```text
address,private_key
0123abcd...,0x...
```

## Build

The build links OpenSSL's secp256k1 implementation and includes the Ethereum
Keccak-256 implementation in `deps/keccak`.

```sh
cd .
zig build -Doptimize=ReleaseFast
zig build run -Doptimize=ReleaseFast
```

For automation, pass the count directly and answer the mode prompts through
standard input:

```sh
zig build run -Doptimize=ReleaseFast -- --count 1000
```

Use `--worker N` to choose a custom pool size. For example:

```sh
zig build run -Doptimize=ReleaseFast -- --count 10000 --worker 400
```

The worker count is independent of the vanity difficulty. More workers do not
make a difficult pattern mathematically easier, and very high values can add
scheduling overhead on small machines.

The progress bar is written to stderr: green shows completed wallets and red
shows remaining work. Private keys are generated from `std.crypto.random` and
must be kept private.
