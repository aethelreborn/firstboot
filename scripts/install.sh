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
bin_dir="${PREFIX:-$HOME/.local/bin}"
install -Dm755 "$tmp/firstboot" "$bin_dir/firstboot"

if [ -z "${PREFIX:-}" ]; then
  case ":${PATH}:" in
    *:"$bin_dir":*) ;;
    *)
      for profile in "$HOME/.profile" "$HOME/.bashrc"; do
        [ -f "$profile" ] || : > "$profile"
        if ! grep -Fqx 'export PATH="$HOME/.local/bin:$PATH"' "$profile" 2>/dev/null; then
          printf '\n# Added by firstboot installer\nexport PATH="$HOME/.local/bin:$PATH"\n' >> "$profile"
        fi
      done
      ;;
  esac
fi

echo "Installed firstboot to $bin_dir/firstboot"
case ":${PATH}:" in
  *:"$bin_dir":*) echo "Run: firstboot" ;;
  *) echo "Run now: $bin_dir/firstboot"; echo "Then open a new shell or run: export PATH=\"$bin_dir:\$PATH\"" ;;
esac
