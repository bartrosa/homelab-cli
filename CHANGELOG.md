# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added — v0.3.0 "Developer Environment"

- **`lab stack`** (rename from `lab toolchain`; aliases `toolchain`, `tc`): component registry, dependency-ordered install, presets, GPU detection, shell PATH management (`shellrc`).
- **26 stack components** across language, build-tool, container, GPU, VCS, package-manager, and database-embedded categories (Python/Node/Go/Rust/Scala/Kotlin, uv, CUDA/ROCm, DuckDB/SQLite, Docker/Podman, …). No Ruby, no .NET.
- **`lab services` framework**: compose orchestrator, shared `homelab-net`, embedded templates, stdin prompt engine, secret handling, service presets.
- **15 services**: postgres (pgvector/postgis/timescaledb plugins), mysql/mariadb, redis, valkey, mongodb, clickhouse, rabbitmq, nats, qdrant, weaviate, prometheus, grafana (auto-provisioned datasources), loki, tempo, minio.
- **High-level wrappers**: `lab obs up/down`, `lab vector up`, `lab data up`.
- **Config**: `stack.*` and extended `services.*` (runtime `auto`, instances, presets).
- **Docs**: [`docs/services.md`](docs/services.md), updated README/commands/configuration.
- **Tests**: stack orchestrator/presets/shellrc/gpu, services orchestrator/templates/postgres, prompt stdin.

### Changed

- `lab toolchain` → `lab stack` (alias preserved).
- `services.runtime` default: `auto` (prefer docker on Ubuntu, podman on Silverblue).

## [0.2.0] - TBD

**Provisioning Release** — distribute `lab`, create verified bootable USB installers, and bootstrap Ubuntu or Fedora Silverblue on a fresh OS.

Implementation landed iteratively in v0.1.0–v0.1.1; v0.2.0 is the documented stable milestone for the full workflow.

### Added

- Release plumbing: GoReleaser `.tar.gz`, `.deb`, `.rpm`, and `checksums.txt` for linux/darwin amd64/arm64.
- `scripts/install.sh` and `scripts/uninstall.sh` — curl install with SHA256 verification, prefix detection, optional PATH setup.
- `lab self-update` with `--check`, `--version`, `--pre-release`, and `--yes` (`internal/updater`).
- `lab iso list|download|disks|images|write` — ISO catalog, verified downloads (SHA256 + GPG), USB disk listing, safe `dd` writes on Linux (`internal/iso`).
- `lab bootstrap essentials` — idempotent baseline for Ubuntu and Fedora Silverblue (`internal/bootstrap`, `internal/pkgmgr`: apt, rpm-ostree, detect).
- Testable process runner: `internal/exec`.
- Provisioning guide: [`docs/provisioning.md`](docs/provisioning.md).

### Changed

- README: v0.2.0 provisioning walkthrough, install/upgrade docs, full command status table.
- `docs/commands.md`: ISO interactive write, `bootstrap essentials` flags, `self-update` reference.
- `docs/architecture.md`: adapter model, `internal/iso` and `internal/updater` packages.
- Terminal UX (carried from v0.1.1): progress bars, GPG key import, theme-adaptive output, interactive ISO picker.
- CI: GitHub Actions Node 24 runtime; golangci-lint v2.12.2 for Go 1.25.

## [0.1.1] - 2026-07-15

### Changed

- Terminal output: theme-adaptive colors, bold/underlined table headers, structured `lab version` and `lab iso list|disks` layout (`--no-color` for plain text).
- `lab iso download`: progress bar with speed and ETA, spinner while connecting, spinner during SHA256/GPG verification.
- `lab iso download`: import upstream GPG signing keys before verification (Ubuntu, Fedora); prominent security banner on GPG failure.
- `lab iso disks`: fix empty device list (missing `TYPE` column from lsblk).
- `lab iso write`: interactive picker, `lab iso write ubuntu-desktop --usb`, `--device` alias for `--to`; new `lab iso images`.
- `lab iso write`: fix lsblk device path (`/dev/sda` not `sda`); auto-unmount USB partitions before burn.
- `lab iso write`: parse lsblk sizes with locale decimal comma (e.g. `58,6G`).
- `lab iso write`: auto `sudo` for dd/sync; umount mounted partitions by mount path (not whole disk).
- `lab iso write`: progress bar while burning (parse dd status); unified yellow `!` warnings (no emoji).
- `lab iso write`: progress polls `/sys/block/*/stat` every 250ms (smooth bar, not dd stderr).
- `scripts/install.sh`: automatically configure shell PATH when installing to a non-standard prefix; `--no-path` to opt out.
- `scripts/uninstall.sh`: remove homelab-cli PATH block from shell rc on uninstall.
- CI: upgrade GitHub Actions to Node 24 runtime; bump golangci-lint for Go 1.25 compatibility.

## [0.1.0] - 2026-07-15

### Added

- Release plumbing: GoReleaser `.deb`, `.rpm`, `.tar.gz`, and `checksums.txt` for linux/darwin amd64/arm64.
- `scripts/install.sh` and `scripts/uninstall.sh` for curl-based installs with SHA256 verification.
- `lab self-update` with `--check`, `--version`, `--pre-release`, and `--yes`.
- `lab iso list|download|disks|write` — ISO catalog, verified downloads, USB disk listing, and safe `dd` writes (Linux).
- `lab bootstrap essentials` for Ubuntu and Fedora Silverblue with idempotent sections.
- Package manager adapters: `internal/pkgmgr` (`apt`, `rpm-ostree`, detect).
- Testable exec runner: `internal/exec`.
- **Server:** `lab server run`, `lab server deploy` (sync, provision, compose, full).
- **PostgreSQL:** `lab postgres apply --config` — idempotent users/databases from YAML.
- **Bare metal:** `lab baremetal install` for qdrant, milvus, clickhouse (Linux).
- **System:** `lab system usb list` and `lab system usb`.
- **Services:** `lab services ensure` for ml-stack.
- **Media:** `lab media heic`.
- Cobra/Viper CLI, bootstrap profiles, `pkg`, `toolchain`, `repos backup`, `ssh`, `templates`.

### Changed

- GoReleaser config: nfpms, changelog filters, release mode.
- GitHub release workflow unchanged in trigger but documents full artifact set.
- `lab ssh sync` and deploy use `internal/server` instead of homelab shell scripts.

### Removed

- `lab media playlist` and YouTube/yt-dlp code.
- `media.*` config keys (`cookies_browser`, `cookies_file`, `downloads_dir`).
