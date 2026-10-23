#!/usr/bin/env sh
set -eu

repo="aethelreborn/firstboot"
tmp=""

# Never make a piped installer disappear on terminals that do not interpret ANSI.
if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
  esc=$(printf '\033')
  blue="${esc}[38;5;81m"; cyan="${esc}[36m"; green="${esc}[32m"
  yellow="${esc}[33m"; red="${esc}[31m"; bold="${esc}[1m"
  dim="${esc}[2m"; reset="${esc}[0m"
else
  blue=""; cyan=""; green=""; yellow=""; red=""; bold=""; dim=""; reset=""
fi

cleanup() {
  status=$?
  [ -n "$tmp" ] && rm -rf "$tmp"
  if [ "$status" -ne 0 ]; then
    printf '\n%s\n' "${red}firstboot installer failed (exit $status). Run again with: NO_COLOR=1 sh -x${reset}" >&2
  fi
  exit "$status"
}
trap cleanup EXIT INT TERM

printf '\n%s\n' "${blue}${bold}FIRSTBOOT INSTALLER${reset}"
printf '%s\n' "${dim}A fresh OS deserves a proper firstboot.${reset}"
printf '%s\n\n' "${dim}This will download and install the latest Firstboot release.${reset}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$os" in
  darwin) os="darwin" ;;
  linux) os="linux" ;;
  *) printf '%s\n' "${red}ERROR: unsupported OS: $os${reset}" >&2; exit 1 ;;
esac
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) printf '%s\n' "${red}ERROR: unsupported architecture: $arch${reset}" >&2; exit 1 ;;
esac

printf '%s' "${cyan}[1/4] Detecting platform...${reset}"
sleep 0.15
printf '\r%s\n' "${green}[1/4] Platform: $os/$arch${reset}"

printf '%s' "${cyan}[2/4] Checking latest release...${reset}"
version="$(curl -fsSL --retry 3 "https://api.github.com/repos/$repo/releases/latest" | grep -o '"tag_name"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"//; s/"$//')"
if [ -z "$version" ]; then
  printf '\r%s\n' "${red}[2/4] No GitHub release found.${reset}" >&2
  exit 1
fi
printf '\r%s\n' "${green}[2/4] Release: $version${reset}"

tmp="$(mktemp -d)"
archive="$tmp/firstboot.tar.gz"
url="https://github.com/$repo/releases/download/$version/firstboot_${os}_${arch}.tar.gz"

printf '%s\n' "${yellow}[3/4] Downloading binary — curl will show live speed and ETA:${reset}"
# Keep this download on stderr-visible progress output while the archive goes to disk.
curl -fL --retry 3 --progress-bar --output "$archive" "$url"
printf '%s\n' "${green}[3/4] Download complete${reset}"

printf '%s' "${cyan}[4/4] Extracting and installing...${reset}"
tar -xzf "$archive" -C "$tmp"
if [ -n "${PREFIX:-}" ]; then
  bin_dir="$PREFIX"
else
  bin_dir=""
  old_ifs="$IFS"
  IFS=:
  for path_dir in $PATH; do
    if [ -n "$path_dir" ] && [ -d "$path_dir" ] && [ -w "$path_dir" ]; then
      bin_dir="$path_dir"
      break
    fi
  done
  IFS="$old_ifs"
  [ -n "$bin_dir" ] || bin_dir="$HOME/.local/bin"
fi
mkdir -p "$bin_dir"
install -m 755 "$tmp/firstboot" "$bin_dir/firstboot"
printf '\r%s\n' "${green}[4/4] Installed to $bin_dir/firstboot${reset}"

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

printf '\n%s\n' "${green}${bold}DONE: Firstboot is installed.${reset}"
case ":${PATH}:" in
  *:"$bin_dir":*) printf '%s\n' "Run it with: firstboot" ;;
  *) printf '%s\n' "Run it now with: $bin_dir/firstboot"; printf '%s\n' "For future shells: source ~/.bashrc"; printf '%s\n' "Note: a piped child shell cannot refresh its parent shell's PATH." ;;
esac
