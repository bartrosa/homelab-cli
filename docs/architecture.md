# Architecture

`lab` is a thin orchestration layer: Cobra commands call into small internal packages that will grow **adapters** for external tools. This PR wires the skeleton only; adapters arrive in focused follow-ups.

## High-level flow

```
┌─────────────┐   persistent flags   ┌──────────────┐
│  Cobra CLI  │ ───────────────────► │ Viper config │
│  (cmd tree) │                      └──────┬───────┘
└──────┬──────┘                             │
       │ PersistentPreRunE                │
       ▼                                   │
┌──────────────┐     stderr/json/text ┌───▼────────────┐
│ slog.Logger  │ ◄────────────────────│ logging.New() │
└──────┬───────┘                      └────────────────┘
       │
       ▼
┌───────────────────────────────────────────────┐
│ Command RunE (stubs today, real work later)   │
└───────────────────────────────────────────────┘
```

## Planned adapters (not implemented in PR #1)

| Domain | Planned abstraction | Backing tools |
|--------|----------------------|---------------|
| Packages | `Packager` interface + OS autodetection | `brew`, `apt`, `dnf`, `pacman`, … |
| Toolchains | `ToolchainRunner` | `mise` (install/use/list) |
| Services | `StackRunner` | compose templates + `docker`/`podman` |
| Repos | `Provider` + local git | `go-git` + GitHub/GitLab/Gitea HTTP APIs |
| Cluster | `ClusterClient` | `k3s` install scripts + `client-go` |
| Models / ML | thin wrappers | `ollama`, `vLLM`, HF cache helpers |
| MCP | stdio server | subset of `lab` commands exposed as MCP tools |

## Package layout conventions

- `cmd/lab` — `main` only.
- `internal/cli` — root command, wiring, shared execution helpers.
- `internal/cli/commands` — individual command constructors (stubs).
- `internal/clierrors` — sentinel errors shared without import cycles.
- `internal/config`, `internal/logging`, `internal/buildinfo` — cross-cutting utilities.
- `pkg/` — reserved for libraries that may become reusable outside this binary.

## MCP direction

`lab mcp serve` will eventually expose a stdio MCP server so Cursor/VS Code can call curated, read-only or guarded operations (list hosts, validate config, dry-run plans). That work is intentionally deferred until the underlying commands are real.
