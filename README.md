# homelab-cli

[![CI](https://github.com/bartrosa/homelab-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/bartrosa/homelab-cli/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.25+-00ADD8.svg)](https://go.dev/dl/)

**`lab`** is a single CLI for homelab automation: bootstrap laptops and servers, manage toolchains and compose stacks, sync your [homelab](https://github.com/bartrosa/homelab) repo to remote hosts, provision PostgreSQL, download media, and create bootable USB installers — with orchestration in Go and thin wrappers around standard system tools.

Module: `github.com/bartrosa/homelab-cli`

## Design principles

- **Orchestration in Go** — workflows, config, retries, and terminal UI live in this repo.
- **No homelab shell/Python scripts from `lab`** — logic migrated here; the personal homelab repo remains the source of compose files, YAML, and project templates.
- **External binaries where required** — `ssh`, `podman-compose`, `wget`, `dd`, etc. See [`docs/external-binaries.md`](docs/external-binaries.md).

## What works today

| Area | Commands | Notes |
|------|----------|--------|
| **Bootstrap** | `bootstrap laptop\|server\|profile\|list\|essentials` | Profiles + `essentials` for Ubuntu/Silverblue |
| **Packages** | `pkg install\|ensure\|list` | brew, apt, dnf, rpm-ostree |
| **Toolchains** | `toolchain install\|list\|use` | via [mise](https://mise.jdx.dev/) |
| **Services** | `services up\|down\|list\|logs\|ensure` | homelab `ml-stack` compose |
| **Server** | `server run`, `server deploy` | SSH + rsync; deploy can provision PG and start compose |
| **PostgreSQL** | `postgres apply` | Idempotent apply from `instances.yaml` (pgx) |
| **Bare metal** | `baremetal install` | Qdrant, Milvus, ClickHouse on Linux |
| **System** | `system usb list`, `system usb` | Bootable USB; ISOs discovered from Ubuntu/Fedora mirrors |
| **ISO** | `iso list`, `iso download`, `iso disks`, `iso write` | Cache ISOs, list disks, burn USB (Linux) |
| **SSH** | `ssh connect`, `ssh sync` | Host inventory from config |
| **Repos** | `repos backup` | GitLab account mirror (homelab Python script today) |
| **Templates** | `templates list\|new` | Copy `project-initiators/` from homelab |
| **Media** | `media heic` | HEIC→JPEG via `heif-convert` |
| **Meta** | `version`, `self-update` | Build metadata and in-place upgrades |

**Planned (stubs):** `cluster`, `gpu`, `models`, `mlops`, `vector`, `pipelines`, `agents`, `obs`, `logs`, `mcp`, most of `repos` beyond backup.

Full command tables: [`docs/commands.md`](docs/commands.md).

## Installation

### One-liner install

```bash
curl -sSL https://raw.githubusercontent.com/bartrosa/homelab-cli/main/scripts/install.sh | bash
```

Pin a version or install to `$HOME/.local`:

```bash
curl -sSL https://raw.githubusercontent.com/bartrosa/homelab-cli/main/scripts/install.sh | bash -s -- --version v0.2.0
curl -sSL https://raw.githubusercontent.com/bartrosa/homelab-cli/main/scripts/install.sh | bash -s -- --prefix "$HOME/.local"
```

### Upgrading

```bash
lab self-update
lab self-update --check   # exit 0 if current, 3 if update available
```

### Alternatives

**Build from source**

```bash
git clone https://github.com/bartrosa/homelab-cli.git
cd homelab-cli
make install    # or: make build && ./bin/lab
```

Requires **Go 1.25+**.

**go install**

```bash
go install github.com/bartrosa/homelab-cli/cmd/lab@latest
```

**Release packages**

Tagged releases publish `.tar.gz`, `.deb`, and `.rpm` via GoReleaser on the [Releases](https://github.com/bartrosa/homelab-cli/releases) page.

## Provisioning a new machine

```bash
# On an existing Linux host:
lab iso list
lab iso download ubuntu-desktop
lab iso disks
lab iso write ~/.cache/homelab-cli/iso/ubuntu-24.04.3-desktop-amd64.iso --to /dev/sdb

# After booting the fresh OS and installing lab:
lab bootstrap essentials
```

## Quick start

1. Copy and edit config:

```bash
mkdir -p ~/.config/homelab-cli
cp docs/config.example.yaml ~/.config/homelab-cli/config.yaml
# set homelab.root to your homelab repo path
```

2. Preview bootstrap, install tools, run stacks:

```bash
lab bootstrap laptop --dry-run
lab pkg ensure ripgrep jq git
lab toolchain install go rust python
lab services list
lab services up ml-stack
```

3. Remote server (set `server.*` in config):

```bash
lab ssh sync
lab server deploy provision    # rsync + postgres apply (local, against PG in YAML)
lab server deploy compose      # rsync + podman-compose on server
lab services ensure            # ml-stack up + service URLs
```

4. USB installer:

```bash
lab system usb list
lab system usb --distro ubuntu-lts-24.04 --device /dev/sdb --workdir ~/Downloads
```

## Global flags

| Flag | Description |
|------|-------------|
| `--config` | Config file (default `~/.config/homelab-cli/config.yaml`) |
| `--homelab-root` | Override `homelab.root` / `LAB_HOMELAB_ROOT` |
| `--dry-run` | Print planned external commands without running them |
| `--no-color` | Disable lipgloss styling |
| `--log-level` | `debug\|info\|warn\|error` |
| `--log-format` | `text\|json` |

Environment variables use the `LAB_` prefix (e.g. `LAB_SERVER_HOST`, `LAB_HOMELAB_ROOT`).

## Configuration

Precedence: **CLI flags → `LAB_*` env → YAML → defaults**.

| Key | Purpose |
|-----|---------|
| `homelab.root` | Path to personal homelab repo (compose, templates, postgres config) |
| `server.host`, `server.user`, `server.port`, `server.path` | Default remote host for rsync/SSH |
| `ssh.hosts` | Named SSH targets for `lab ssh connect` |
| `services.runtime` | `podman-compose` or `docker` |
| `repos.providers` | Git hosting tokens for future clone/backup |

Details: [`docs/configuration.md`](docs/configuration.md) · example: [`docs/config.example.yaml`](docs/config.example.yaml).

## Documentation

| Document | Description |
|----------|-------------|
| [`docs/commands.md`](docs/commands.md) | Command reference with status |
| [`docs/configuration.md`](docs/configuration.md) | Config keys and precedence |
| [`docs/architecture.md`](docs/architecture.md) | Packages and data flow |
| [`docs/external-binaries.md`](docs/external-binaries.md) | Required host tools |
| [`docs/homelab-migration.md`](docs/homelab-migration.md) | homelab repo → `lab` migration map |
| [`CHANGELOG.md`](CHANGELOG.md) | Release notes |

## Relationship to the homelab repo

**homelab-cli** is the productized CLI. The **homelab** git repo is the “scratchpad”: compose stacks, postgres `instances.yaml`, `project-initiators/`, and docs. Point `homelab.root` at that checkout so `lab` can find compose files and templates. New automation should land in Go here; homelab scripts are retired as features migrate.

## Development

```bash
make ci    # fmt, vet, lint, test, build
```

See [`CONTRIBUTING.md`](CONTRIBUTING.md).

## License

Apache-2.0 — see [`LICENSE`](LICENSE).
