#!/usr/bin/env bash
set -euo pipefail

ZIG_VERSION="${ZIG_VERSION:-0.13.0}"
INSTALL_ROOT="${HOME}/.local/zig"
ZIG_DIR="${INSTALL_ROOT}/zig-linux-x86_64-${ZIG_VERSION}"

if [[ "$(uname -s)" != "Linux" ]]; then
    printf 'This installer currently supports Linux only.\n' >&2
    exit 1
fi

case "$(uname -m)" in
    x86_64) ZIG_ARCH="x86_64" ;;
    aarch64|arm64) ZIG_ARCH="aarch64"; ZIG_DIR="${INSTALL_ROOT}/zig-linux-aarch64-${ZIG_VERSION}" ;;
    *) printf 'Unsupported architecture: %s\n' "$(uname -m)" >&2; exit 1 ;;
esac

install_packages() {
    local packages=(curl tar xz-utils build-essential libssl-dev ca-certificates)
    if [[ "$(id -u)" -eq 0 ]]; then
        apt-get update
        apt-get install -y "${packages[@]}"
    elif command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
        sudo apt-get update
        sudo apt-get install -y "${packages[@]}"
    else
        printf 'Missing packages. Run this once with an account that has sudo access:\n'
        printf '  sudo apt-get update && sudo apt-get install -y curl tar xz-utils build-essential libssl-dev ca-certificates\n'
        exit 1
    fi
}

if ! command -v cc >/dev/null 2>&1 || ! dpkg-query -W -f='${Status}' libssl-dev 2>/dev/null | grep -q 'install ok installed'; then
    install_packages
fi

mkdir -p "${INSTALL_ROOT}"
if [[ ! -x "${ZIG_DIR}/zig" ]]; then
    archive="/tmp/zig-${ZIG_VERSION}.tar.xz"
    url="https://ziglang.org/download/${ZIG_VERSION}/zig-linux-${ZIG_ARCH}-${ZIG_VERSION}.tar.xz"
    printf 'Downloading Zig %s for %s...\n' "${ZIG_VERSION}" "${ZIG_ARCH}"
    curl --fail --location --retry 3 --output "${archive}" "${url}"
    tar -xJf "${archive}" -C "${INSTALL_ROOT}"
    rm -f "${archive}"
fi

ln -sfn "${ZIG_DIR}" "${INSTALL_ROOT}/current"
export PATH="${INSTALL_ROOT}/current:${PATH}"

shell_file="${HOME}/.bashrc"
if [[ "${SHELL:-}" == */zsh ]]; then shell_file="${HOME}/.zshrc"; fi
path_line='export PATH="$HOME/.local/zig/current:$PATH"'
if [[ -f "${shell_file}" ]] && ! grep -Fqx "${path_line}" "${shell_file}"; then
    printf '\n%s\n' "${path_line}" >> "${shell_file}"
elif [[ ! -f "${shell_file}" ]]; then
    printf '%s\n' "${path_line}" > "${shell_file}"
fi

printf 'Zig %s installed. Building evm...\n' "$(zig version)"
zig build -Doptimize=ReleaseFast
printf '\nSetup complete. Run: ./run.sh\n'
