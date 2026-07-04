#!/bin/sh
# Installs the latest leanapi release for your OS/architecture.
#
# Usage:
#   curl --proto '=https' --tlsv1.2 -sSf https://raw.githubusercontent.com/dylanjnsn02/leanapi/main/install.sh | sh
#
# Installs into $HOME/.local/bin by default (no sudo required). Override with:
#   curl ... | INSTALL_DIR=/usr/local/bin sh

set -eu

REPO="dylanjnsn02/leanapi"
BIN_NAME="leanapi"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

say() { printf '%s\n' "$1"; }
err() { printf 'error: %s\n' "$1" >&2; exit 1; }

detect_platform() {
  os="$(uname -s)"
  arch="$(uname -m)"

  case "$os" in
    Darwin) os="darwin" ;;
    Linux) os="linux" ;;
    *) err "unsupported OS: $os -- see https://github.com/$REPO/releases for manual downloads" ;;
  esac

  case "$arch" in
    x86_64 | amd64) arch="amd64" ;;
    arm64 | aarch64) arch="arm64" ;;
    *) err "unsupported architecture: $arch" ;;
  esac

  printf '%s-%s\n' "$os" "$arch"
}

fetch() {
  url="$1"
  out="$2"
  if command -v curl >/dev/null 2>&1; then
    curl --proto '=https' --tlsv1.2 -sSfL "$url" -o "$out"
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$url" -O "$out"
  else
    err "curl or wget is required to install leanapi"
  fi
}

main() {
  platform="$(detect_platform)"
  asset="${BIN_NAME}-${platform}"
  url="https://github.com/${REPO}/releases/latest/download/${asset}"

  tmp="$(mktemp "${TMPDIR:-/tmp}/${BIN_NAME}.XXXXXX")"
  trap 'rm -f "$tmp"' EXIT

  say "Downloading ${asset} from the latest release..."
  fetch "$url" "$tmp" || err "download failed: $url (is there a release for $platform yet?)"

  chmod +x "$tmp"

  mkdir -p "$INSTALL_DIR"
  dest="${INSTALL_DIR}/${BIN_NAME}"
  mv "$tmp" "$dest"

  say "Installed leanapi to $dest"

  case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *)
      say ""
      say "$INSTALL_DIR is not on your PATH. Add this to your shell profile (e.g. ~/.zshrc or ~/.bashrc):"
      say "  export PATH=\"$INSTALL_DIR:\$PATH\""
      ;;
  esac

  say ""
  say "Run 'leanapi' to get started."
}

main
