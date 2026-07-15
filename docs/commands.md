# Command reference

> **Legend:** ✅ implemented · 🚧 scaffolded (`not implemented yet`)

**v0.3.0 (Developer Environment):** `lab stack`, full `lab services` framework, and high-level `lab obs` / `lab vector` / `lab data` wrappers.

**v0.2.0 (Provisioning Release):** install script, `lab self-update`, `lab iso *`, and `lab bootstrap essentials` are ✅ ready. See [`provisioning.md`](provisioning.md) for the full new-machine workflow.

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

| Flag | Description |
|------|-------------|
| `--target` | `auto` (default), `ubuntu`, or `silverblue` |
| `--yes` | Non-interactive; accept defaults |
| `--skip` | Comma-separated sections to skip (e.g. `docker,mise`) |
| `--only` | Comma-separated sections to run (e.g. `cli-basics,build`) |
| `--dry-run` | Print planned commands without executing (global flag) |

Sections: `system-update`, `cli-basics`, `shell-tools`, `build`, `container-runtime`, `mise`, `distrobox`, `flatpak-flathub`.

```bash
lab bootstrap laptop --dry-run
lab bootstrap essentials --dry-run --target silverblue
lab bootstrap essentials --only cli-basics --yes
lab bootstrap profile dgx-spark   # from config bootstrap.profiles
```

### `lab iso`

| Command | Description | Status |
|---------|-------------|--------|
| `lab iso list` | Supported distros and resolved versions. | ✅ |
| `lab iso download <distro>` | Download and verify ISO to cache (GPG + SHA256). | ✅ |
| `lab iso images` | List cached ISO files ready to burn. | ✅ |
| `lab iso disks` | List block devices; USB vs SYSTEM (Linux). | ✅ |
| `lab iso write [distro\|path]` | Burn ISO with safety checks (Linux). Interactive if no args. | ✅ |

Flow: `list` → `download` → `disks` → `write`.

```bash
lab iso download ubuntu-desktop
lab iso images
lab iso write                              # interactive: pick cached ISO + USB drive
lab iso write ubuntu-desktop --usb         # burn by distro name / cached alias
lab iso write ~/.cache/homelab-cli/iso/ubuntu-24.04.3-desktop-amd64.iso --to /dev/sdb
lab iso write ubuntu-desktop --device sda  # --device is alias for --to
```

On non-Linux platforms, `disks` and `write` return `not implemented yet` (stubs).

**Note:** `lab system usb` is an alternate path that discovers Ubuntu/Fedora images from release metadata and writes in one step. Prefer `lab iso` for the verified download + cache + burn workflow.

### `lab pkg`

| Command | Description | Status |
|---------|-------------|--------|
| `lab pkg install <name> [more...]` | Install packages (brew / apt / dnf / rpm-ostree). | ✅ |
| `lab pkg ensure <name> [more...]` | Install only if missing. | ✅ |
| `lab pkg list` | Show common packages and detected backend. | ✅ |

### `lab stack` (alias: `lab toolchain`, `lab tc`)

Developer environment components: languages, build tools, containers, GPU stacks, embedded databases.

| Command | Description | Status |
|---------|-------------|--------|
| `lab stack list [--category]` | List available components by category. | ✅ |
| `lab stack list-installed` | Show installed components and versions. | ✅ |
| `lab stack info <component>` | Component details (backend, requires, PATH entries). | ✅ |
| `lab stack gpu` | Detect GPUs and suggest compute stacks. | ✅ |
| `lab stack install <component>...` | Install components (dependency order, skip if installed). | ✅ |
| `lab stack install --preset <name> [--gpu]` | Install a preset bundle. | ✅ |
| `lab stack preset list` | List stack presets. | ✅ |
| `lab stack preset show <name>` | Show components in a preset. | ✅ |
| `lab stack path` | Print managed shell PATH block. | ✅ |
| `lab stack path refresh` | Regenerate PATH block in shell rc. | ✅ |
| `lab stack path remove` | Remove managed block. | ✅ |
| `lab stack use <lang> <version>` | Activate mise version (legacy). | ✅ |

| Flag | Description |
|------|-------------|
| `--yes` | Non-interactive |
| `--dry-run` | Print plan only |
| `--force` | Reinstall / override GPU checks |
| `--skip-path` | Skip shell rc update |
| `--component-version` | Override version for install |

```bash
lab stack install --preset ml --yes
lab stack install postgres duckdb sqlite --yes   # embedded DBs in stack, not services
lab stack install rust --yes
lab stack install --preset gpu-nvidia --yes      # adds CUDA when NVIDIA GPU detected
lab toolchain install python --yes               # alias works
```

### `lab services`

Compose-backed local services on shared network `homelab-net`. Config under `~/.config/homelab-cli/services/<id>/`.

| Command | Description | Status |
|---------|-------------|--------|
| `lab services list [--category]` | List services and running/stopped status. | ✅ |
| `lab services info <id>` | Description and config schema. | ✅ |
| `lab services init <id>` | Interactive wizard or `--set` / `--yes`. | ✅ |
| `lab services up <id> [more...]` | Start service(s) or `--preset`. | ✅ |
| `lab services down <id> [more...]` | Stop service(s). | ✅ |
| `lab services restart <id>` | Restart one service. | ✅ |
| `lab services status [id]` | Runtime status. | ✅ |
| `lab services logs <id> [-f] [--tail N]` | Tail logs (compose). | 🚧 partial |
| `lab services connect <id> [--interactive]` | Print connection string or open CLI client. | ✅ |
| `lab services rm <id> [--data]` | Remove config (+ data with `--data`). | ✅ |
| `lab services preset list` | List service presets. | ✅ |
| `lab services preset show <name>` | Show services in preset. | ✅ |
| `lab services ensure [ml-stack]` | Legacy homelab-repo ml-stack compose. | ✅ |

| Flag | Description |
|------|-------------|
| `--yes` | Non-interactive init |
| `--set key=value` | Config override (repeatable; comma lists for multiselect) |
| `--preset` | Preset name for init/up |
| `--force` | Overwrite existing init |

```bash
lab services init postgres --set plugins=pgvector,postgis --set expose=local --yes
lab services up postgres
lab services up --preset observability --yes
lab services up --preset ml-stack --yes
lab services connect postgres --interactive
```

High-level wrappers: `lab obs up`, `lab vector up qdrant`, `lab data up postgres`.

See [`services.md`](services.md) for all 15 services, fields, and connection examples.

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
| `lab self-update` | Install latest release from GitHub. | ✅ |

| Flag | Command | Description |
|------|---------|-------------|
| `--check` | `self-update` | Exit 0 if current, 3 if update available, 1 on error |
| `--version <tag>` | `self-update` | Force install specific release (downgrade allowed) |
| `--pre-release` | `self-update` | Include GitHub prereleases |
| `--yes` | `self-update` | Skip confirmation prompt |

If the installed binary is not writable (e.g. `/usr/local/bin/lab`), `self-update` prints instructions to re-run with `sudo` — it does not escalate privileges automatically.

```bash
lab version --output json
lab self-update --check
lab self-update --yes
lab --log-level debug services list
```
