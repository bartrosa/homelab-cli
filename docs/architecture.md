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
│ Domain packages (bootstrap, iso, server, …)   │
│  → exec.Runner / executil (dry-run, logging)  │
│  → ssh, wget, podman-compose, dd, gpg, …     │
└───────────────────────────────────────────────┘
```

## Adapter model (target)

Commands stay thin; reusable logic lives in `internal/*` adapters:

| Adapter | Interface role | Implementations (current / planned) |
|---------|------------------|-------------------------------------|
| **Package managers** | Install/ensure/list OS packages | ✅ brew, apt, dnf, rpm-ostree (`internal/packager`, `internal/pkgmgr`) |
| **Toolchain** | Language runtime install/switch | ✅ mise wrapper (`internal/toolchain`) |
| **Services** | Compose stack lifecycle | ✅ podman-compose / docker (`internal/services`, `internal/mlstack`) |
| **Repos** | Clone, backup, sync, status | ✅ GitLab backup (interim Python); 🚧 go-git + REST providers |
| **Cluster** | k3s/k8s ops | 🚧 kubectl / client-go |
| **Storage** | S3-compatible ops | 🚧 MinIO API client |
| **Models / ML** | Local LLM pull/run | 🚧 ollama, vLLM wrappers |

Stub commands (`commands.StubRunE`) reserve the CLI surface until each adapter ships — typically one adapter per PR.

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
| `internal/system` | Bootable USB (mirror discovery) | HTTP, wget, sha256sum, dd |
| `internal/iso` | ISO catalog, download, verify, burn | gpg, wget/curl, dd, lsblk |
| `internal/baremetal` | DB installers on Linux | curl, apt, sudo, systemd |
| `internal/updater` | In-place binary upgrade | GitHub releases API |

## Cross-cutting

| Package | Role |
|---------|------|
| `internal/config` | YAML + `LAB_*` env |
| `internal/homelabroot` | Resolve homelab repo path |
| `internal/exec`, `internal/executil` | Testable command runner with dry-run |
| `internal/ui` | lipgloss sections, tables, progress |
| `internal/platform` | OS detection (brew vs apt vs …) |
| `internal/logging` | slog on context |
| `internal/buildinfo` | Version ldflags |
| `internal/clierrors` | Shared errors (`ErrNotImplemented`, …) |

## Command groups (Cobra)

| Group ID | Commands |
|----------|----------|
| `foundation` | bootstrap, pkg, toolchain, services, **iso** |
| `repos` | repos |
| `infra` | server, postgres, baremetal, system, ssh, cluster, gpu, containers, net, storage |
| `data` | models, data, notebooks, mlops, vector, pipelines, agents |
| `workflow` | obs, logs, templates, media, mcp |
| `meta` | version, **self-update** |

Run `lab --help` to see groups in the root help output.

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
| Repos | Go GitLab/GitHub API + go-git instead of Python backup |
| Bootstrap | Replace `script:` steps with native Go where practical |
| Cluster / GPU / net | Thin adapters over kubectl, nvidia-smi, tailscale CLI |
| Models / ML | ollama/vLLM pull, MLflow status |
| MCP | `lab mcp serve` as stdio server exposing a guarded subset of read-only tools for Cursor/Copilot |

## Layout conventions

- `cmd/lab` — `main` only (signal-aware context, exit code 1 on error)
- `internal/cli` — root command wiring, config/logger bootstrap
- `internal/cli/commands` — one file per domain; `NewXxxCmd()` constructors
- `internal/cli/appctx` — `Session` (config, dry-run, styles) on `context.Context`
- `pkg/` — reserved for future public libraries (e.g. shared repo provider interfaces)
