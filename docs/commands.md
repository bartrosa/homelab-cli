# Command reference

> Status legend: ✅ implemented · 🚧 scaffolded (`not implemented yet`)

## Foundation — bootstrap & install

| Command | Description | Status |
|---------|-------------|--------|
| `lab bootstrap laptop` | Full laptop bootstrap (packages, shell, fonts, dotfiles, git, containers). | 🚧 |
| `lab bootstrap server` | Server bootstrap (SSH baseline, fail2ban-style hardening, agents). | 🚧 |
| `lab bootstrap profile <name>` | Bootstrap using `bootstrap.profiles.<name>` from config. | 🚧 |
| `lab pkg install <name>` | Install via native package manager abstraction. | 🚧 |
| `lab pkg ensure <name>` | Idempotent install/upgrade guardrail. | 🚧 |
| `lab pkg list` | List packages tracked by `lab`. | 🚧 |
| `lab toolchain install <lang> [more]` | Install toolchains via `mise`. | 🚧 |
| `lab toolchain list` | Show installed toolchains/versions. | 🚧 |
| `lab toolchain use <lang> <version>` | Activate a toolchain version. | 🚧 |
| `lab services up <name> [more]` | Start compose stacks (postgres, redis, …). | 🚧 |
| `lab services down <name> [more]` | Stop stacks. | 🚧 |
| `lab services list` | List stacks + status. | 🚧 |
| `lab services logs <name>` | Tail stack logs. | 🚧 |

## Repos — multi-repo management

| Command | Description | Status |
|---------|-------------|--------|
| `lab repos clone <pattern>` | Clone org/user patterns across providers. | 🚧 |
| `lab repos backup` | Mirror configured remotes to disk/object storage. | 🚧 |
| `lab repos sync` | Fetch/pull all managed clones. | 🚧 |
| `lab repos status` | Dirty / ahead-behind overview. | 🚧 |
| `lab repos list` | List remotes visible to configured providers. | 🚧 |

## Infra & networking

| Command | Description | Status |
|---------|-------------|--------|
| `lab cluster status` | Cluster health summary. | 🚧 |
| `lab cluster kubeconfig` | kubeconfig helpers for homelab contexts. | 🚧 |
| `lab gpu info` | GPU/driver diagnostics. | 🚧 |
| `lab ssh connect <host>` | SSH helper with inventory + keys. | 🚧 |
| `lab containers ps` | Cross-runtime container listing. | 🚧 |
| `lab net status` | Tailscale/WireGuard/DNS/mDNS snapshot. | 🚧 |
| `lab storage ls <uri>` | S3-compatible listing helper. | 🚧 |

## Data / AI / ML / MLOps

| Command | Description | Status |
|---------|-------------|--------|
| `lab models pull <name>` | Pull/cache local LLMs. | 🚧 |
| `lab data sync` | Dataset sync helpers (DVC/lakeFS/etc.). | 🚧 |
| `lab notebooks up` | Launch notebook servers. | 🚧 |
| `lab mlops status` | Experiment tracker connectivity. | 🚧 |
| `lab vector list` | Vector DB inventory/status. | 🚧 |
| `lab pipelines run <name>` | Local pipeline runner entrypoint. | 🚧 |
| `lab agents list` | Local agent runtime registry. | 🚧 |

## Workflow & observability

| Command | Description | Status |
|---------|-------------|--------|
| `lab obs up` | Start observability bundle. | 🚧 |
| `lab logs tail <selector>` | Aggregate log tailing. | 🚧 |
| `lab templates new <name>` | Generate projects from templates. | 🚧 |
| `lab mcp serve` | stdio MCP server for IDE integrations. | 🚧 |

## Meta

| Command | Description | Status |
|---------|-------------|--------|
| `lab version` | Print build metadata (`--output text|json`). | ✅ |

Examples:

```bash
lab version --output json
lab --log-level debug --log-format json version
```
