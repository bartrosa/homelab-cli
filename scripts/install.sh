#!/bin/sh
set -eu

# homelab-cli install script (POSIX sh)
# Usage:
#   curl -sSL https://raw.githubusercontent.com/bartrosa/homelab-cli/main/scripts/install.sh | sh
#   curl -sSL ... | sh -s -- --version v0.1.0
#   curl -sSL ... | sh -s -- --prefix "$HOME/.local"

REPO="bartrosa/homelab-cli"
PROJECT="homelab-cli"
VERSION=""
PREFIX=""
FORCE=0
CHECK=0
NO_PATH=0

PATH_MARKER="# homelab-cli: PATH"

log() { printf '%s\n' "$*" >&2; }
die() { log "error: $*"; exit 1; }

while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    --prefix) PREFIX="$2"; shift 2 ;;
    --force) FORCE=1; shift ;;
    --check) CHECK=1; shift ;;
    --no-path) NO_PATH=1; shift ;;
    -h|--help)
      log "Usage: install.sh [--version TAG] [--prefix PATH] [--force] [--check] [--no-path]"
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

detect_shell_rc() {
  case "${SHELL:-}" in
    */zsh)
      printf '%s' "$HOME/.zshrc"
      ;;
    */bash)
      printf '%s' "$HOME/.bashrc"
      ;;
    *)
      if [ -f "$HOME/.bashrc" ]; then
        printf '%s' "$HOME/.bashrc"
      elif [ -f "$HOME/.profile" ]; then
        printf '%s' "$HOME/.profile"
      else
        printf '%s' "$HOME/.profile"
      fi
      ;;
  esac
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

path_configured_in_rc() {
  rc=$1
  bindir=$2
  [ -f "$rc" ] || return 1
  grep -qF "$PATH_MARKER" "$rc" 2>/dev/null && return 0
  grep -qF "$bindir" "$rc" 2>/dev/null
}

ensure_path_in_rc() {
  bindir=$1

  if in_path "$bindir"; then
    log "info: $bindir already in PATH for this session"
    return 0
  fi

  rc=$(detect_shell_rc)
  line="export PATH=\"$bindir:\$PATH\""

  if path_configured_in_rc "$rc" "$bindir"; then
    log "info: PATH already configured in $rc"
    log "info: run: source $rc   (or open a new terminal)"
    return 0
  fi

  mkdir -p "$(dirname "$rc")"
  {
    printf '\n%s\n' "$PATH_MARKER"
    printf '%s\n' "$line"
  } >> "$rc"

  log "info: added $bindir to PATH in $rc"
  log "info: run: source $rc   (or open a new terminal)"
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
    if [ "$NO_PATH" -eq 0 ] && ! in_path "$bindir"; then
      log "[check] would append $bindir to $(detect_shell_rc)"
    fi
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

  if [ "$NO_PATH" -eq 0 ]; then
    ensure_path_in_rc "$bindir"
  fi

  log ""
  log "Installed successfully."
  log "  lab version"
  log "  lab --help"
  log "  lab self-update"
}

main "$@"
