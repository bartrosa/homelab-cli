# homelab-cli

[![CI](https://github.com/bartrosa/homelab-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/bartrosa/homelab-cli/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.25+-00ADD8.svg)](https://go.dev/dl/)

**`lab`** — CLI for end-to-end homelab automation: from bare metal to GPU-served LLMs.

One binary for bootstrapping machines, language toolchains, compose-backed services, multi-repo workflows, remote deploys, and bootable USB installers — orchestration in Go, thin wrappers around standard system tools.

Module: `github.com/bartrosa/homelab-cli` · Latest release: **v0.2.0**

## Why?

Homelab work spans dozens of domains: OS packages, runtimes, databases, Git mirrors, SSH deploys, ISO provisioning, and (eventually) clusters and local LLMs. Without a single entry point, automation drifts into scattered shell scripts under `~/scripts` with no shared config, logging, or dry-run semantics.

`lab` centralizes that:

- **One CLI** with grouped commands, `--help` everywhere, and consistent flags.
- **Declarative config** (`~/.config/homelab-cli/config.yaml` + `LAB_*` env) instead of hard-coded paths.
- **Orchestration in Go** — retries, terminal UI, idempotent steps — while delegating to `mise`, `apt`, `podman-compose`, `ssh`, `dd`, etc. where appropriate.
- **Homelab repo as data** — compose files, postgres YAML, and project templates stay in your personal [homelab](https://github.com/bartrosa/homelab) checkout; `lab` reads them via `homelab.root`.

## What can it do?

| Area | Summary |
|------|---------|
| **Bootstrap & install** | Laptop/server profiles, essentials for Ubuntu/Silverblue, packages, developer stack (`lab stack`), compose services, verified ISO download and USB burn. |
| **Stack** | Install languages (Python, Go, Rust, Scala, …), build tools, GPU stacks (CUDA/ROCm), embedded DBs, and managed shell PATH. |
| **Services** | Init and run 17 compose-backed services (Postgres, Redis, observability, vector/graph DBs, MinIO) on shared `homelab-net`. |
| **Repos** | GitLab account backup today; clone/sync/status planned. |
| **Infra & networking** | SSH connect/sync, remote server deploy, PostgreSQL apply, bare-metal DB installers, USB/ISO provisioning. Cluster/GPU/net/storage planned. |
| **Data / AI / ML** | Stubs for models, notebooks, MLOps, vector DBs, pipelines, agents — vector DB install exists under `baremetal install` today. |
| **Observability** | Stubs for obs/logs; services logs work via compose. |
| **Workflow** | Project templates from homelab initiators, HEIC conversion, version/self-update; MCP server planned. |

**Implemented today:** see the [command status table](#command-status) below. Everything else returns `not implemented yet` until the matching adapter lands.

## Design principles

- **Orchestration in Go** — workflows, config, retries, and terminal UI live in this repo.
- **No homelab shell/Python from `lab`** for migrated features; homelab remains the source of compose, YAML, and templates.
- **External binaries where required** — `ssh`, `podman-compose`, `wget`, `dd`, etc. See [`docs/external-binaries.md`](docs/external-binaries.md).

## Installation

### One-liner install

```bash
curl -sSL https://raw.githubusercontent.com/bartrosa/homelab-cli/main/scripts/install.sh | bash
```

The install script downloads the release tarball, verifies SHA256 checksums, and installs `lab`. If `~/.local/bin` is used (when `/usr/local/bin` is not writable), it **automatically appends** the install directory to your shell rc (`~/.bashrc`, `~/.zshrc`, or `~/.profile`). Open a new terminal or run `source ~/.bashrc` afterward.

Pin a version or install to a custom prefix:

```bash
curl -sSL https://raw.githubusercontent.com/bartrosa/homelab-cli/main/scripts/install.sh | bash -s -- --version v0.2.0
curl -sSL https://raw.githubusercontent.com/bartrosa/homelab-cli/main/scripts/install.sh | bash -s -- --prefix "$HOME/.local"
curl -sSL .../install.sh | bash -s -- --no-path   # skip automatic PATH setup
curl -sSL .../install.sh | bash -s -- --check     # dry run — print planned actions
```

After install:

```bash
lab version
lab --help
lab self-update --check
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

Tagged releases publish `.tar.gz`, `.deb`, and `.rpm` via GoReleaser on the [Releases](https://github.com/bartrosa/homelab-cli/releases) page (linux/darwin, amd64/arm64, plus `checksums.txt`).

## Provisioning a new machine (v0.2.0)

Full walkthrough: [`docs/provisioning.md`](docs/provisioning.md).

On an **existing Linux host** — download a verified ISO and burn a USB drive:

```bash
lab iso list
lab iso download ubuntu-desktop
lab iso disks
lab iso write ubuntu-desktop --usb
```

Boot the target machine from USB, install the OS, then on the **fresh install**:

```bash
curl -sSL https://raw.githubusercontent.com/bartrosa/homelab-cli/main/scripts/install.sh | bash
lab bootstrap essentials --dry-run
lab bootstrap essentials --yes
lab stack install --preset backend --yes
lab stack install rust cmake --yes
source ~/.bashrc
```

Supported ISO resolvers today: **ubuntu-desktop**, **fedora-silverblue** (plus catalog stubs for debian, arch, nixos, …). `lab bootstrap essentials` targets **Ubuntu** and **Fedora Silverblue** (`--target auto|ubuntu|silverblue`).

## Setting up your dev environment

After `lab bootstrap essentials`, install a curated developer stack:

```bash
lab stack install --preset backend --yes    # git, docker, python, node, uv, go, make
lab stack install rust cmake --yes          # ad-hoc components
lab stack list                              # all components by category
lab stack list-installed                    # what's on this machine
lab stack path refresh                      # update ~/.bashrc managed PATH block
source ~/.bashrc
```

Presets include `minimal`, `basic`, `backend`, `frontend`, `systems`, `jvm`, `ml`, `data`, `gpu-nvidia`, `gpu-amd`, and `full`. Override versions in config (`stack.components`) or with `--component-version`.

`lab toolchain` is an alias for `lab stack` (same commands).

## Running services

Local data and observability stacks run via Docker/Podman Compose on `homelab-net`:

```bash
lab services init postgres --set plugins=pgvector,postgis --set expose=local --yes
lab services up postgres
lab services connect postgres                 # connection string
lab services connect postgres --interactive # psql session

lab services up --preset observability --yes  # prometheus, loki, tempo, grafana
lab services up --preset ml-stack --yes       # postgres, qdrant, minio, clickhouse

# Graph databases + GraphRAG stack
lab services init arcadedb --set databases=knowledge_graph --yes
lab services up arcadedb
lab services connect arcadedb
lab services up --preset graphrag --yes       # arcadedb + qdrant + minio + postgres
lab services up --preset graph-lab --yes      # arcadedb + nebulagraph side by side
lab vector up arcadedb                        # graph DB with built-in vector search

lab obs up                                    # wrapper → observability preset
lab vector up qdrant
lab data up postgres
```

Non-interactive / secrets from 1Password:

```bash
op run --env-file=.op.env -- lab services up --preset ml-stack --yes
```

Service catalog: [`docs/services.md`](docs/services.md). Legacy homelab-repo ml-stack: `lab services ensure ml-stack`.

## Quick start

Realistic examples across domains (✅ = works today, 🚧 = stub):

```bash
# Bootstrap a fresh machine (after OS install)
lab bootstrap laptop                    # ✅ profile-based setup
lab bootstrap essentials                # ✅ Ubuntu or Silverblue baseline packages
lab stack install --preset ml --yes     # ✅ developer stack (alias: lab toolchain)
lab services init postgres --yes        # ✅ local compose services
lab services up --preset observability  # ✅ prometheus, grafana, loki, tempo

# Provision boot media (Linux)
lab iso list
lab iso download ubuntu-desktop
lab iso disks
lab iso write                           # ✅ interactive picker
lab iso write ubuntu-desktop --usb      # ✅ burn by distro name

# Remote homelab server (set server.* in config)
lab ssh sync
lab server deploy full                  # ✅ rsync + postgres apply + compose

# Planned — returns "not implemented yet"
lab repos clone "github.com/me/*"       # 🚧
lab models pull llama3                  # 🚧
lab cluster status                      # 🚧
```

### First-time config

```bash
mkdir -p ~/.config/homelab-cli
cp docs/config.example.yaml ~/.config/homelab-cli/config.yaml
# set homelab.root to your homelab repo path
```

### Common workflows

```bash
lab bootstrap laptop --dry-run
lab pkg ensure ripgrep jq git
lab stack install go rust python        # alias: lab toolchain install
lab services list
lab services up postgres
lab postgres apply --config ~/homelab/postgres/config/instances.yaml
lab templates new golang ~/projects/my-api
lab version --output json
```

Full provisioning flow for a new machine: [`docs/commands.md`](docs/commands.md#lab-iso).

## Command status

Legend: ✅ ready · 🚧 planned (stub — `not implemented yet`)

### Foundation — bootstrap & install

| Command | Status |
|---------|--------|
| `lab bootstrap laptop` | ✅ |
| `lab bootstrap server` | ✅ |
| `lab bootstrap profile <name>` | ✅ |
| `lab bootstrap list` | ✅ |
| `lab bootstrap essentials` | ✅ |
| `lab pkg install <name>` | ✅ |
| `lab pkg ensure <name>` | ✅ |
| `lab pkg list` | ✅ |
| `lab stack install <component>` | ✅ |
| `lab stack list` | ✅ |
| `lab stack install --preset <name>` | ✅ |
| `lab stack path refresh` | ✅ |
| `lab toolchain …` | ✅ (alias for `lab stack`) |
| `lab services init <id>` | ✅ |
| `lab services up <id>` | ✅ |
| `lab services down <id>` | ✅ |
| `lab services list` | ✅ |
| `lab services connect <id>` | ✅ |
| `lab services up --preset <name>` | ✅ |
| `lab services ensure` | ✅ (legacy homelab ml-stack) |
| `lab obs up` | ✅ |
| `lab vector up <id>` | ✅ |
| `lab data up <id>` | ✅ |
| `lab iso list \| download \| disks \| images \| write` | ✅ |

### Repos — multi-repo management

| Command | Status |
|---------|--------|
| `lab repos clone <pattern>` | 🚧 |
| `lab repos backup` | ✅ |
| `lab repos sync` | 🚧 |
| `lab repos status` | 🚧 |
| `lab repos list` | 🚧 |

### Infra & networking

| Command | Status |
|---------|--------|
| `lab cluster status`, `lab cluster kubeconfig` | 🚧 |
| `lab gpu info` | 🚧 |
| `lab ssh connect`, `lab ssh sync` | ✅ |
| `lab containers ps` | 🚧 |
| `lab net status` | 🚧 |
| `lab storage ls` | 🚧 |
| `lab server run`, `lab server deploy` | ✅ |
| `lab postgres apply` | ✅ |
| `lab baremetal install` | ✅ |
| `lab system usb list`, `lab system usb` | ✅ |

### Data / AI / ML / MLOps

| Command | Status |
|---------|--------|
| `lab models pull` | 🚧 |
| `lab data sync` | 🚧 |
| `lab notebooks up` | 🚧 |
| `lab mlops status` | 🚧 |
| `lab vector list` | 🚧 |
| `lab pipelines run` | 🚧 |
| `lab agents list` | 🚧 |

### Workflow & observability

| Command | Status |
|---------|--------|
| `lab obs up` | 🚧 |
| `lab logs tail` | 🚧 |
| `lab templates list \| new` | ✅ |
| `lab media heic` | ✅ |
| `lab mcp serve` | 🚧 |
| `lab version` | ✅ |
| `lab self-update` | ✅ |

Per-command flags and examples: [`docs/commands.md`](docs/commands.md).

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
| `repos.providers` | Git hosting tokens for clone/backup |

Details: [`docs/configuration.md`](docs/configuration.md) · example: [`docs/config.example.yaml`](docs/config.example.yaml).

## Documentation

| Document | Description |
|----------|-------------|
| [`docs/provisioning.md`](docs/provisioning.md) | **v0.2.0** new-machine workflow (install → ISO → USB → essentials) |
| [`docs/commands.md`](docs/commands.md) | Command reference with status and examples |
| [`docs/configuration.md`](docs/configuration.md) | Config keys and precedence |
| [`docs/architecture.md`](docs/architecture.md) | Packages, adapters, and data flow |
| [`docs/external-binaries.md`](docs/external-binaries.md) | Required host tools |
| [`docs/homelab-migration.md`](docs/homelab-migration.md) | homelab repo → `lab` migration map |
| [`CHANGELOG.md`](CHANGELOG.md) | Release notes (current: v0.1.1; next: v0.2.0) |

## Relationship to the homelab repo

**homelab-cli** is the productized CLI. The **homelab** git repo is the “scratchpad”: compose stacks, postgres `instances.yaml`, `project-initiators/`, and docs. Point `homelab.root` at that checkout so `lab` can find compose files and templates. New automation should land in Go here; homelab scripts are retired as features migrate.

## Development

```bash
make ci    # fmt, vet, lint, test, build
```

See [`CONTRIBUTING.md`](CONTRIBUTING.md).

## License

Apache-2.0 — see [`LICENSE`](LICENSE).
