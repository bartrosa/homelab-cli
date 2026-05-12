# homelab-cli

[![CI](https://github.com/OWNER/REPO/actions/workflows/ci.yml/badge.svg)](https://github.com/OWNER/REPO/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.23.0+-00ADD8.svg)](https://go.dev/dl/)

> Replace `OWNER/REPO` in badge URLs after you publish the repository.

**Module path placeholder:** replace `__MODULE_PATH__` in `go.mod`, imports, and docs with your real module path (for example `github.com/you/homelab-cli`).

CLI for end-to-end homelab automation — from bare metal to GPU-served LLMs.

## Why?

Homelab automation tends to sprawl across ad-hoc shell scripts, README fragments, and one-off Ansible snippets. `lab` is a single entry point that will grow into a declaratively configured toolkit: one binary, consistent UX, and clear seams for adapters (package managers, compose stacks, Git providers, cluster clients).

## What can it do?

The command tree is grouped into eight areas. **Everything is scaffolded today** except `lab version`, which prints build metadata.

1. **Bootstrap** — laptop/server profiles, baseline packages, dotfiles, and hardening.
2. **Toolchains** — language runtimes via `mise` wrappers (Go, Node, Bun, Deno, Python, Rust, BEAM, Zig, Java, Ruby, …).
3. **Services** — local Postgres/Redis/Mongo/Kafka/RabbitMQ/MinIO/ClickHouse/etcd/NATS stacks via compose.
4. **Repos** — bulk clone/mirror/backup across GitHub, GitLab, Gitea, and Codeberg.
5. **Cluster & net** — k3s/k8s helpers, GPU diagnostics, SSH inventory, Tailscale/WireGuard.
6. **Data / AI / ML** — models, datasets, notebooks, MLOps, vector DBs, local pipelines, agents.
7. **Observability** — Prometheus/Grafana/Loki/Tempo bundles plus aggregated logs.
8. **Workflow** — project templates, optional MCP stdio server for IDE integrations.

## Installation

### Build from source

```bash
git clone https://github.com/OWNER/REPO.git
cd homelab-cli
make install   # or: make build && ./bin/lab
```

### `go install`

```bash
go install __MODULE_PATH__/cmd/lab@latest
```

### GitHub Releases

After the first tagged release, prefer the checksum-verified archives published by GoReleaser. A convenience installer script lives at `scripts/install.sh` (TODO until release artifacts exist).

## Quick start

```bash
lab bootstrap laptop                    # set up a fresh machine (stub)
lab toolchain install go bun rust       # install language toolchains (stub)
lab services up postgres redis          # spin up databases (stub)
lab repos clone "github.com/me/*"       # clone all your repos (stub)
lab models pull llama3                  # pull a local LLM (stub)
lab cluster status                      # check homelab k3s (stub)
lab version                             # ✅ prints build info
```

## Commands

See [`docs/commands.md`](docs/commands.md) for the full reference. Summary:

| Group | Commands | Status |
|------|----------|--------|
| Foundation | `bootstrap`, `pkg`, `toolchain`, `services` | 🚧 planned |
| Repos | `repos` | 🚧 planned |
| Infra | `cluster`, `gpu`, `ssh`, `containers`, `net`, `storage` | 🚧 planned |
| Data / AI / ML | `models`, `data`, `notebooks`, `mlops`, `vector`, `pipelines`, `agents` | 🚧 planned |
| Workflow | `obs`, `logs`, `templates`, `mcp` | 🚧 planned |
| Meta | `version` | ✅ ready |

## Configuration

Precedence: **CLI flags → environment (`LAB_*`) → YAML file → defaults**.

Default config path: `~/.config/homelab-cli/config.yaml`.

Example:

```yaml
log_level: info
log_format: text

bootstrap:
  default_profile: default
  profiles: {}

repos:
  root: ~/src
  backup_dir: ~/backups/repos
  providers:
    - name: github-personal
      kind: github
      host: github.com
      token_env: GH_TOKEN

services:
  stacks_dir: ~/.config/homelab-cli/stacks
  runtime: podman

cluster:
  kubeconfig: ~/.kube/config
  context: ""

storage:
  endpoint: ""
  access_key: ""
```

More detail: [`docs/configuration.md`](docs/configuration.md).

## Development

```bash
make ci
```

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for conventions, tooling versions, and PR expectations.

## Go version note

`go.mod` currently declares `go 1.23.0` so `golangci-lint` releases (built with older toolchains) can analyze the module without tripping over `go 1.25` language gates. You can still compile with Go 1.25+ locally. When golangci-lint ships binaries built with Go ≥1.25, bump the `go` directive to match your target.

## License

Apache-2.0 — see [`LICENSE`](LICENSE).
