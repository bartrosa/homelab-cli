#!/bin/sh
set -eu

# homelab-cli install script (POSIX sh)
# Usage:
#   curl -sSL https://raw.githubusercontent.com/bartrosa/homelab-cli/main/scripts/install.sh | sh
#   curl -sSL ... | sh -s -- --version v0.2.0
#   curl -sSL ... | sh -s -- --prefix "$HOME/.local"

REPO="bartrosa/homelab-cli"
PROJECT="homelab-cli"
VERSION=""
PREFIX=""
FORCE=0
CHECK=0

log() { printf '%s\n' "$*" >&2; }
die() { log "error: $*"; exit 1; }

while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    --prefix) PREFIX="$2"; shift 2 ;;
    --force) FORCE=1; shift ;;
    --check) CHECK=1; shift ;;
    -h|--help)
      log "Usage: install.sh [--version TAG] [--prefix PATH] [--force] [--check]"
      exit 0
      ;;
    *) die "unknown argument: $1" ;;
  esac
done

detect_os() {
  uname -s | tr '[:upper:]' '[:lower:]'
}

detect_arch() {
  m=$(uname -m)
  case "$m" in
    x86_64|amd64) printf '%s' "amd64" ;;
    aarch64|arm64) printf '%s' "arm64" ;;
    *) die "unsupported architecture: $m" ;;
  esac
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

writable_dir() {
  dir=$1
  [ -d "$dir" ] || mkdir -p "$dir" 2>/dev/null || return 1
  touch "$dir/.write-test" 2>/dev/null || return 1
  rm -f "$dir/.write-test"
  return 0
}

choose_prefix() {
  if [ -n "$PREFIX" ]; then
    printf '%s' "$PREFIX"
    return
  fi
  if writable_dir "/usr/local/bin"; then
    printf '%s' "/usr/local"
    return
  fi
  home=${HOME:-}
  [ -n "$home" ] || die "HOME not set and /usr/local/bin not writable"
  log "info: /usr/local/bin not writable — installing to $home/.local/bin"
  log "info: ensure $home/.local/bin is in your PATH"
  printf '%s' "$home/.local"
}

fetch_latest_version() {
  need_cmd curl
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | head -n1 \
    | cut -d '"' -f4
}

in_path() {
  case ":${PATH:-}:" in
    *:"$1":*) return 0 ;;
    *) return 1 ;;
  esac
}

main() {
  need_cmd curl
  need_cmd tar
  need_cmd sha256sum

  os=$(detect_os)
  arch=$(detect_arch)
  case "$os" in
    linux|darwin) ;;
    *) die "unsupported OS: $os" ;;
  esac

  tag=${VERSION:-}
  if [ -z "$tag" ]; then
    tag=$(fetch_latest_version)
  fi
  [ -n "$tag" ] || die "could not determine release version"

  prefix=$(choose_prefix)
  bindir="$prefix/bin"
  dest="$bindir/lab"

	ver="${tag#v}"
	asset="${PROJECT}_${ver}_${os}_${arch}.tar.gz"
	base="https://github.com/${REPO}/releases/download/${tag}"
  url="${base}/${asset}"
  checksums_url="${base}/checksums.txt"

  log "install: $tag → $dest ($os/$arch)"

  if [ "$CHECK" -eq 1 ]; then
    log "[check] would download $url"
    log "[check] would verify with $checksums_url"
    log "[check] would install to $dest"
    exit 0
  fi

  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT INT HUP

  curl -fsSL "$url" -o "$tmp/$asset"
  curl -fsSL "$checksums_url" -o "$tmp/checksums.txt"

  hash=$(grep " ${asset}$" "$tmp/checksums.txt" | awk '{print $1}')
  [ -n "$hash" ] || die "checksum entry not found for $asset"
  got=$(sha256sum "$tmp/$asset" | awk '{print $1}')
  [ "$got" = "$hash" ] || die "checksum mismatch"

  tar -xzf "$tmp/$asset" -C "$tmp" lab

  mkdir -p "$bindir"
  if [ -f "$dest" ] && [ "$FORCE" -eq 0 ]; then
    die "$dest already exists (use --force to overwrite)"
  fi

  if [ "$prefix" = "/usr/local" ] || [ "$prefix" = "/usr" ]; then
    if ! writable_dir "$bindir"; then
      need_cmd sudo
      sudo install -m 0755 "$tmp/lab" "$dest"
    else
      install -m 0755 "$tmp/lab" "$dest"
    fi
  else
    install -m 0755 "$tmp/lab" "$dest"
  fi

  log ""
  log "Next steps:"
  log "  lab version"
  log "  lab --help"
  log "  lab self-update"

  if ! in_path "$bindir"; then
    log ""
    log "Add to your shell rc:"
    log "  export PATH=\"$bindir:\$PATH\""
  fi
}

main "$@"
