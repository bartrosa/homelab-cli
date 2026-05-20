# homelab → homelab-cli migration

Reference: the personal **homelab** repo (compose, YAML, templates, notes) vs **homelab-cli** (`lab`).

**Legend:** ✅ in CLI (Go orchestration) · 🔶 partial · ⬜ not migrated · 📁 stays in homelab · 🔧 needs external binary ([`external-binaries.md`](external-binaries.md))

## Principles

1. **Orchestration in Go** — steps, config, retries, UI.
2. **No homelab `.sh` / `.py` from `lab`** for migrated features.
3. **Allowed `exec`** — standard host tools listed in [`external-binaries.md`](external-binaries.md).

---

## `tools/`

| homelab | lab | Notes |
|---------|-----|--------|
| `tools/media/yt_playlist_download.sh` | 📁 | Stay in homelab; not in `lab` |
| `tools/media/heic_converter.sh` | ✅ `lab media heic` | Go + heif-convert |
| `tools/dev-setup/*.sh` | ✅ `lab pkg`, `lab toolchain`, bootstrap profiles | YAML profiles |
| `tools/gitlab/backup_account.py` | 🔶 `lab repos backup` | Still exec Python |
| `tools/system/bootable_usb/` | ✅ `lab system usb` | Dynamic mirror query + wget/dd |
| `tools/cad/*.md` | 📁 | Research notes |

---

## `scripts/`

| homelab | lab | Notes |
|---------|-----|--------|
| `sync-to-server.sh` | ✅ `lab ssh sync`, `lab server deploy` | rsync + ssh |
| `remote-run.sh` | ✅ `lab server run` | ssh |
| `deploy-and-compose.sh` | ✅ `lab server deploy [provision\|compose\|full]` | PG apply local via pgx |
| `ensure-ml-stack-up.sh` | ✅ `lab services ensure` | podman-compose |
| `install-*-bare-metal.sh` | ✅ `lab baremetal install <name>` | Go + curl/apt/sudo |
| `install-server-deps.sh` | 🔶 bootstrap `script:` step | Still bash from homelab |
| `run-ml-stack-setup-on-server.sh` | ⬜ | Not in bootstrap profiles yet |
| `build-langfuse-image.sh`, `test-mlflow-api-route.sh` | ⬜ | |
| `git-newbranch.sh`, `setup-git-hooks.sh` | ⬜ | |

---

## `postgres/`, `ml-stack/`

| homelab | lab | Notes |
|---------|-----|--------|
| `postgres/provision` (Python) | ✅ `lab postgres apply` | pgx |
| `postgres/config/instances.yaml` | 📁 | Path under homelab checkout |
| `ml-stack/podman-compose.yml` | ✅ `lab services up ml-stack` | Compose from homelab.root |

---

## Config in homelab-cli

| Area | Status |
|------|--------|
| `homelab.root`, `LAB_HOMELAB_ROOT`, `--homelab-root` | ✅ |
| `server.*` (deploy, sync, run) | ✅ |
| `ssh.hosts`, `lab ssh connect` | ✅ |
| Bootstrap embedded + custom profiles | ✅ 🔶 ML server setup script not wired |
| `repos.backup` | ✅ 🔶 Python |
| `media heic`, `services`, `postgres`, `baremetal`, `system usb` | ✅ |
| cluster, gpu, models, mlops, mcp, … | ⬜ stubs |

---

## Open decisions

| Area | Options |
|------|---------|
| `lab repos backup` | Port to Go (GitLab API + go-git) vs keep Python |
| Bootstrap `script:` steps | Keep one-off homelab bash vs rewrite in Go |
| `run-ml-stack-setup-on-server.sh` | New `lab server setup` or bootstrap profile |
