# Command reference

> **Legend:** ✅ implemented · 🚧 scaffolded (`not implemented yet`)

Run `lab <command> --help` for flags. Global flags apply to all commands: `--config`, `--homelab-root`, `--dry-run`, `--no-color`, `--log-level`, `--log-format`.

---

## Foundation — bootstrap & install

### `lab bootstrap`

| Command | Description | Status |
|---------|-------------|--------|
| `lab bootstrap laptop` | Built-in laptop profile (macOS / Linux / Silverblue). | ✅ |
| `lab bootstrap server` | Ubuntu server profile + optional homelab `install-server-deps.sh` step. | ✅ |
| `lab bootstrap profile <name>` | Run built-in or config-defined profile. | ✅ |
| `lab bootstrap list` | List embedded profiles. | ✅ |
| `lab bootstrap essentials` | Baseline packages for Ubuntu or Silverblue (sections: system-update, cli-basics, …). | ✅ |

Built-in profiles: `laptop-macos`, `laptop-linux`, `silverblue-laptop`, `server-ubuntu`.

```bash
lab bootstrap laptop --dry-run
lab bootstrap essentials --dry-run --target silverblue
lab bootstrap profile dgx-spark   # from config bootstrap.profiles
```

### `lab iso`

| Command | Description | Status |
|---------|-------------|--------|
| `lab iso list` | Supported distros and resolved versions. | ✅ |
| `lab iso download <distro>` | Download and verify ISO to cache. | ✅ |
| `lab iso disks` | List block devices; USB vs SYSTEM (Linux). | ✅ |
| `lab iso write <iso> --to <device>` | Burn ISO with safety checks (Linux). | ✅ |

Flow: `list` → `download` → `disks` → `write`.

```bash
lab iso download ubuntu-desktop
lab iso write ~/.cache/homelab-cli/iso/ubuntu-24.04.3-desktop-amd64.iso --to /dev/sdb
```

### `lab pkg`

| Command | Description | Status |
|---------|-------------|--------|
| `lab pkg install <name> [more...]` | Install packages (brew / apt / dnf / rpm-ostree). | ✅ |
| `lab pkg ensure <name> [more...]` | Install only if missing. | ✅ |
| `lab pkg list` | Show common packages and detected backend. | ✅ |

### `lab toolchain`

| Command | Description | Status |
|---------|-------------|--------|
| `lab toolchain install <lang> [more...]` | Install runtimes via `mise`. | ✅ |
| `lab toolchain list` | List installed toolchains. | ✅ |
| `lab toolchain use <lang> <version>` | Activate a version. | ✅ |

### `lab services`

Manages compose stacks under `homelab.root` (e.g. `ml-stack/podman-compose.yml`). Runtime from `services.runtime` (default `podman-compose`).

| Command | Description | Status |
|---------|-------------|--------|
| `lab services up <name> [more...]` | Start stack(s). | ✅ |
| `lab services down <name> [more...]` | Stop stack(s). | ✅ |
| `lab services list` | List stacks and compose paths. | ✅ |
| `lab services logs <name>` | Tail logs. | ✅ |
| `lab services ensure [ml-stack]` | `podman-compose up -d` for ml-stack; print service URLs. | ✅ |

```bash
lab services up ml-stack
lab services ensure
```

---

## Repos — multi-repo management

| Command | Description | Status |
|---------|-------------|--------|
| `lab repos clone <pattern>` | Clone org/user patterns across providers. | 🚧 |
| `lab repos backup` | GitLab account mirror (runs homelab `backup_account.py`). | ✅ |
| `lab repos sync` | Fetch/pull managed clones. | 🚧 |
| `lab repos status` | Dirty / ahead-behind overview. | 🚧 |
| `lab repos list` | List remotes for configured providers. | 🚧 |

Requires `GITLAB_TOKEN` (or provider `token_env`) and `homelab.root` for the backup script path.

---

## Infra & networking

### `lab server`

Uses `server.*` from config (remote homelab checkout path).

| Command | Description | Status |
|---------|-------------|--------|
| `lab server run '<shell command>'` | Run command on server in `server.path`. | ✅ |
| `lab server deploy` | Rsync homelab repo to server only. | ✅ |
| `lab server deploy provision` | Rsync + `lab postgres apply` locally (`postgres/config/instances.yaml`). | ✅ |
| `lab server deploy compose` | Rsync + `podman-compose up -d` in `ml-stack` on server. | ✅ |
| `lab server deploy full` | Provision + compose. | ✅ |

```bash
lab server run 'cd ml-stack && podman-compose ps'
lab server deploy full --dry-run
```

### `lab postgres`

| Command | Description | Status |
|---------|-------------|--------|
| `lab postgres apply --config <path>` | Apply databases/users from YAML (idempotent). | ✅ |

Requires `POSTGRES_ADMIN_PASSWORD` or `PGPASSWORD`. Config format matches homelab `postgres/config/instances.yaml`.

### `lab baremetal`

Run **on the target Linux server** (uses curl, apt, sudo).

| Command | Description | Status |
|---------|-------------|--------|
| `lab baremetal install qdrant` | Qdrant from GitHub release + systemd. | ✅ |
| `lab baremetal install milvus` | Milvus DEB from GitHub. | ✅ |
| `lab baremetal install clickhouse` | ClickHouse from official apt repo. | ✅ |

### `lab system`

| Command | Description | Status |
|---------|-------------|--------|
| `lab system usb list` | Query Ubuntu (meta-release) and Fedora (dl.fedoraproject.org) for desktop/Silverblue ISOs. | ✅ |
| `lab system usb` | Download ISO, verify SHA256, write to block device with `dd`. | ✅ |

**Discovered images (typical):** two recent Ubuntu LTS (≥ 22.04), latest Ubuntu interim, latest Fedora Silverblue.

| Flag | Description |
|------|-------------|
| `--distro` | Image ID from `list` or alias: `ubuntu-latest`, `ubuntu-lts`, `fedora-silverblue`, `ubuntu-24.04`, … |
| `--device` | Block device (e.g. `/dev/sdb`; on macOS often `/dev/diskN`) |
| `--workdir` | Download directory |
| `--iso-url` | Skip discovery; use custom ISO URL |

```bash
lab system usb list
lab system usb --distro ubuntu-25.10 --device /dev/sdb --workdir ~/Downloads
```

### `lab ssh`

| Command | Description | Status |
|---------|-------------|--------|
| `lab ssh connect <alias>` | Interactive SSH to `ssh.hosts.<alias>`. | ✅ |
| `lab ssh sync` | Rsync `homelab.root` to `server.*` (same as deploy sync). | ✅ |

### Other infra (stubs)

| Command | Description | Status |
|---------|-------------|--------|
| `lab cluster status` | Cluster health summary. | 🚧 |
| `lab cluster kubeconfig` | kubeconfig helpers. | 🚧 |
| `lab gpu info` | GPU/driver diagnostics. | 🚧 |
| `lab containers ps` | Container listing. | 🚧 |
| `lab net status` | Tailscale/WireGuard/DNS snapshot. | 🚧 |
| `lab storage ls <uri>` | S3-compatible listing. | 🚧 |

---

## Data / AI / ML / MLOps (stubs)

| Command | Description | Status |
|---------|-------------|--------|
| `lab models pull <name>` | Pull/cache local LLMs. | 🚧 |
| `lab data sync` | Dataset sync helpers. | 🚧 |
| `lab notebooks up` | Notebook servers. | 🚧 |
| `lab mlops status` | Experiment tracker connectivity. | 🚧 |
| `lab vector list` | Vector DB inventory. | 🚧 |
| `lab pipelines run <name>` | Local pipeline runner. | 🚧 |
| `lab agents list` | Agent runtime registry. | 🚧 |

---

## Workflow & observability

| Command | Description | Status |
|---------|-------------|--------|
| `lab templates list` | Template kinds: `golang`, `python`, `rust`, `typescript`. | ✅ |
| `lab templates new <kind> <dir>` | Copy from homelab `project-initiators/`. | ✅ |
| `lab media heic [dir]` | HEIC → JPEG via `heif-convert`. | ✅ |
| `lab obs up` | Observability bundle. | 🚧 |
| `lab logs tail <selector>` | Aggregated logs. | 🚧 |
| `lab mcp serve` | stdio MCP server. | 🚧 |

```bash
lab media heic ~/Pictures/import --quality 95 --force
```

---

## Meta

| Command | Description | Status |
|---------|-------------|--------|
| `lab version` | Build version, commit, date (`--output text\|json`). | ✅ |
| `lab self-update` | Install latest release from GitHub (`--check`, `--version`, `--pre-release`). | ✅ |

```bash
lab version --output json
lab self-update --check
lab --log-level debug services list
```
