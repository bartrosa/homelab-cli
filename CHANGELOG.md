# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Server:** `lab server run`, `lab server deploy` (sync, provision, compose, full) — SSH + rsync; PostgreSQL apply via pgx locally.
- **PostgreSQL:** `lab postgres apply --config` — idempotent users/databases from YAML.
- **Bare metal:** `lab baremetal install` for qdrant, milvus, clickhouse (Linux).
- **System:** `lab system usb list` and `lab system usb` — discover Ubuntu/Fedora ISOs from upstream mirrors; wget, checksum, dd.
- **Services:** `lab services ensure` for ml-stack.
- **Media:** `lab media heic` only (YouTube playlist support removed from CLI).
- **SSH sync:** native rsync (no homelab `sync-to-server.sh`).
- **Docs:** full English documentation refresh; [`docs/README.md`](docs/README.md) index.
- **Dependencies:** `pgx/v5`; Go 1.25 module baseline.

### Removed

- `lab media playlist` and all YouTube/yt-dlp code (`internal/media/playlist*`, `ytdlp.go`).
- `media.*` config keys (`cookies_browser`, `cookies_file`, `downloads_dir`).

### Changed

- `lab ssh sync` and deploy use `internal/server` instead of homelab shell scripts.
- `docs/external-binaries.md`, `docs/homelab-migration.md` updated for current migration state.

### Foundation (earlier unreleased work)

- Cobra/Viper CLI, grouped command tree, `version`, config loader, slog logging, lipgloss UI.
- `bootstrap` profiles: laptop-macos, laptop-linux, silverblue-laptop, server-ubuntu.
- `pkg`, `toolchain` (mise), `services` (homelab compose).
- `repos backup`, `ssh connect`, `templates`, `media heic` (HEIC conversion).
- Global flags: `--dry-run`, `--homelab-root`, `--no-color`.
