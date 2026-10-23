#!/usr/bin/env sh
set -eu

repo="aethelreborn/firstboot"
esc=$(printf '\033')
blue="${esc}[38;5;81m"
cyan="${esc}[36m"
green="${esc}[32m"
yellow="${esc}[33m"
red="${esc}[31m"
bold="${esc}[1m"
dim="${esc}[2m"
reset="${esc}[0m"

cleanup() {
  printf '%s' "${esc}[?25h${reset}"
  [ -n "${tmp:-}" ] && rm -rf "$tmp"
}
trap cleanup EXIT INT TERM
printf '%s' "${esc}[?25l"

printf '\n%s\n' "${blue}${bold}╭──────────────────────────────────────────────╮${reset}"
printf '%s\n' "${blue}${bold}│             FIRSTBOOT INSTALLER              │${reset}"
printf '%s\n\n' "${blue}${bold}╰──────────────────────────────────────────────╯${reset}"
printf '%s\n' "${dim}  A fresh OS deserves a proper firstboot.${reset}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$os" in
  darwin) os="darwin" ;;
  linux) os="linux" ;;
  *) printf '%s\n' "${red}✗ Unsupported OS: $os${reset}" >&2; exit 1 ;;
esac
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) printf '%s\n' "${red}✗ Unsupported architecture: $arch${reset}" >&2; exit 1 ;;
esac

printf '%s' "${cyan}⠋ Detecting platform${reset}"
sleep 0.15
printf '\r%s\n' "${green}✓ Platform: $os/$arch${reset}"

printf '%s' "${cyan}⠋ Checking latest release${reset}"
version="$(curl -fsSL --retry 3 "https://api.github.com/repos/$repo/releases/latest" | sed -n 's/.*"tag_name": "\([^"]*\)".*/\1/p' | head -1)"
if [ -z "$version" ]; then
  printf '\r%s\n' "${red}✗ No release found. Try again later.${reset}" >&2
  exit 1
fi
printf '\r%s\n' "${green}✓ Release $version found${reset}"

tmp="$(mktemp -d)"
archive="$tmp/firstboot.tar.gz"
url="https://github.com/$repo/releases/download/$version/firstboot_${os}_${arch}.tar.gz"

printf '%s\n' "${yellow}↓ Downloading firstboot — live transfer below${reset}"
# curl's progress meter includes percentage, bytes, transfer speed, elapsed time, and ETA.
curl -fL --retry 3 --progress-bar --output "$archive" "$url"
printf '%s\n' "${green}✓ Download complete${reset}"

printf '%s' "${cyan}⠋ Extracting binary${reset}"
tar -xzf "$archive" -C "$tmp"
printf '\r%s\n' "${green}✓ Binary extracted${reset}"

bin_dir="${PREFIX:-$HOME/.local/bin}"
printf '%s' "${cyan}⠋ Installing to $bin_dir${reset}"
mkdir -p "$bin_dir"
install -m 755 "$tmp/firstboot" "$bin_dir/firstboot"
printf '\r%s\n' "${green}✓ Installed firstboot to $bin_dir/firstboot${reset}"

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

printf '\n%s\n' "${green}${bold}✓ Firstboot is ready.${reset}"
case ":${PATH}:" in
  *:"$bin_dir":*) printf '%s\n' "  Run now: ${bold}firstboot${reset}" ;;
  *) printf '%s\n' "  Run now: ${bold}$bin_dir/firstboot${reset}"; printf '%s\n' "  Then run: ${bold}source ~/.bashrc${reset}" ;;
esac
printf '\n'
