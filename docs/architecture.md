# Architecture

`lab` is an orchestration CLI: Cobra commands load config and session context, then call focused packages that plan work and invoke external tools where necessary.

## Request flow

```
┌─────────────┐   persistent flags   ┌──────────────┐
│  Cobra CLI  │ ───────────────────► │ Viper/config │
│  (cmd tree) │                      └──────┬───────┘
└──────┬──────┘                             │
       │ PersistentPreRunE                  │
       ▼                                    │
┌──────────────┐   lipgloss (optional) ┌───▼────────────┐
│ appctx.Session│ ◄───────────────────│ ui.NewStyles() │
│ + slog.Logger │                      └────────────────┘
└──────┬───────┘
       │ RunE
       ▼
┌───────────────────────────────────────────────┐
│ Domain packages (bootstrap, media, server, …) │
│  → executil.Runner (dry-run, logging)         │
│  → exec: ssh, wget, podman-compose, …       │
└───────────────────────────────────────────────┘
```

## Implemented packages

| Package | Role | External tools |
|---------|------|----------------|
| `internal/bootstrap` | Embedded YAML profiles; steps: pkg, toolchain, script | homelab bash for `script:` steps only |
| `internal/packager` | OS package install | brew, apt, dnf, rpm-ostree |
| `internal/toolchain` | Language runtimes | mise |
| `internal/services` | Compose stack up/down/list/logs | podman-compose / docker compose |
| `internal/mlstack` | Ensure ml-stack is up | podman-compose |
| `internal/server` | SSH remote run, rsync deploy | ssh, rsync |
| `internal/postgres` | YAML → PG apply | TCP to PostgreSQL (pgx) |
| `internal/ssh` | Connect + sync | ssh, rsync |
| `internal/repos` | GitLab backup | homelab Python script (interim) |
| `internal/templates` | Project scaffolds | filesystem copy from homelab |
| `internal/media` | HEIC convert | heif-convert |
| `internal/system` | Bootable USB | HTTP to mirrors, wget, sha256sum, dd |
| `internal/baremetal` | DB installers on Linux | curl, apt, sudo, systemd |

## Cross-cutting

| Package | Role |
|---------|------|
| `internal/config` | YAML + `LAB_*` env |
| `internal/homelabroot` | Resolve homelab repo path |
| `internal/executil` | Command runner with dry-run |
| `internal/ui` | lipgloss sections and tables |
| `internal/platform` | OS detection (brew vs apt vs …) |
| `internal/logging` | slog on context |
| `internal/buildinfo` | Version ldflags |

## Command groups (Cobra)

| Group ID | Commands |
|----------|----------|
| `foundation` | bootstrap, pkg, toolchain, services |
| `repos` | repos |
| `infra` | server, postgres, baremetal, system, ssh, cluster, gpu, containers, net, storage |
| `data` | models, data, notebooks, mlops, vector, pipelines, agents |
| `workflow` | obs, logs, templates, media, mcp |
| `meta` | version |

Stub commands return a consistent “not implemented yet” error via `commands.StubRunE`.

## Homelab repo boundary

| Stays in homelab git | Lives in homelab-cli |
|----------------------|----------------------|
| `ml-stack/podman-compose.yml`, `.env.example` | `lab services`, `lab services ensure` |
| `postgres/config/instances.yaml` | `lab postgres apply` |
| `project-initiators/*` | `lab templates new` |
| Docs, experiments, CAD notes | User-facing docs in `docs/` |

Orchestration and new features belong in Go here. See [`homelab-migration.md`](homelab-migration.md).

## Planned work

| Area | Direction |
|------|-----------|
| Repos | Go GitLab API + go-git instead of Python backup |
| Bootstrap | Replace `script:` steps with native Go where practical |
| Cluster / ML | Thin adapters over kubectl, ollama, etc. |
| MCP | `lab mcp serve` exposing guarded read-only tools |

## Layout conventions

- `cmd/lab` — `main` only
- `internal/cli` — root command wiring
- `internal/cli/commands` — per-domain command constructors
- `internal/cli/appctx` — `Session` on `context.Context`
- `pkg/` — reserved for future public libraries
