#!/usr/bin/env bash
set -euo pipefail

if [[ -x "$HOME/.local/zig/current/zig" ]]; then
	export PATH="$HOME/.local/zig/current:$PATH"
fi

zig build run -Doptimize=ReleaseFast -- "$@"