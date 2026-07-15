#!/bin/sh
set -eu

PREFIX="/usr/local"
BIN="$PREFIX/bin/lab"

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

echo "Removed $BIN"
