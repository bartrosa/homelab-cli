#!/bin/sh
set -eu

PREFIX="/usr/local"
BIN="$PREFIX/bin/lab"

PATH_MARKER="# homelab-cli: PATH"

while [ $# -gt 0 ]; do
  case "$1" in
    --prefix)
      PREFIX="$2"
      BIN="$PREFIX/bin/lab"
      shift 2
      ;;
    -h|--help)
      echo "Usage: uninstall.sh [--prefix PATH]" >&2
      exit 0
      ;;
    *) echo "unknown argument: $1" >&2; exit 1 ;;
  esac
done

detect_shell_rc() {
  case "${SHELL:-}" in
    */zsh) printf '%s' "$HOME/.zshrc" ;;
    */bash) printf '%s' "$HOME/.bashrc" ;;
    *)
      if [ -f "$HOME/.bashrc" ]; then printf '%s' "$HOME/.bashrc"
      elif [ -f "$HOME/.profile" ]; then printf '%s' "$HOME/.profile"
      else printf '%s' "$HOME/.profile"
      fi
      ;;
  esac
}

remove_path_from_rc() {
  bindir=$(dirname "$BIN")
  rc=$(detect_shell_rc)
  [ -f "$rc" ] || return 0
  grep -qF "$PATH_MARKER" "$rc" || return 0

  tmp=$(mktemp)
  skip=0
  while IFS= read -r line || [ -n "$line" ]; do
    if [ "$line" = "$PATH_MARKER" ]; then
      skip=1
      continue
    fi
    if [ "$skip" -eq 1 ]; then
      case "$line" in
        *"$bindir"*) skip=0; continue ;;
        *) skip=0 ;;
      esac
    fi
    printf '%s\n' "$line"
  done < "$rc" > "$tmp"
  mv "$tmp" "$rc"
  echo "Removed PATH entry from $rc" >&2
}

if [ ! -f "$BIN" ]; then
  echo "lab not found at $BIN" >&2
  exit 0
fi

if [ -w "$BIN" ]; then
  rm -f "$BIN"
else
  command -v sudo >/dev/null 2>&1 || { echo "need sudo to remove $BIN" >&2; exit 1; }
  sudo rm -f "$BIN"
fi

remove_path_from_rc

echo "Removed $BIN"
