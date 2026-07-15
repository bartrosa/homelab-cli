# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

## [0.1.0] - YYYY-MM-DD

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
