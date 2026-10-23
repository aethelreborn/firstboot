#!/usr/bin/env sh
set -eu
repo="aethelreborn/firstboot"
os="$(uname -s | tr '[:upper:]' '[:lower:]')"; arch="$(uname -m)"
case "$os" in darwin) os="darwin";; linux) os="linux";; *) echo "Unsupported OS: $os" >&2; exit 1;; esac
case "$arch" in x86_64|amd64) arch="amd64";; aarch64|arm64) arch="arm64";; *) echo "Unsupported architecture: $arch" >&2; exit 1;; esac
version="$(curl -fsSL "https://api.github.com/repos/$repo/releases/latest" | sed -n 's/.*"tag_name": "\([^"]*\)".*/\1/p' | head -1)"
[ -n "$version" ] || { echo "No release found; use go run ./cmd/firstboot" >&2; exit 1; }
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
url="https://github.com/$repo/releases/download/$version/firstboot_${os}_${arch}.tar.gz"
curl -fsSL "$url" | tar -xz -C "$tmp"
install -Dm755 "$tmp/firstboot" "${PREFIX:-$HOME/.local/bin}/firstboot"
echo "Installed firstboot to ${PREFIX:-$HOME/.local/bin}/firstboot"
