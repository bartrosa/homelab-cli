# External binaries

**Rule:** workflow logic (ordering, config, retries, parallelism, terminal UI) lives in **Go**.  
**Do not** shell out to `.sh` / `.py` scripts from the homelab repository.

Programs below cannot reasonably be replaced by a Go standard-library call alone. `lab` invokes them with explicit arguments. See also [`homelab-migration.md`](homelab-migration.md).

| Program | Used by | Why not pure Go |
|---------|---------|-----------------|
| **heif-convert** | `lab media heic` | HEIC decoder (libheif) |
| **ssh** | `lab ssh connect`, `lab server run` | Interactive sessions and remote shells |
| **rsync** | `lab ssh sync`, `lab server deploy` | Efficient tree sync to server |
| **podman-compose** / **docker compose** | `lab services`, deploy compose | Container runtime |
| **curl** / **wget** | `lab baremetal install`, `lab system usb` | Large downloads and mirror indexes |
| **apt-get** / **dpkg** | `lab baremetal install` (Milvus, ClickHouse) | Distribution packages on target host |
| **sudo** | bare metal, USB write | Privileged install and `dd` |
| **sha256sum** | `lab system usb` | ISO checksum verification |
| **dd** | `lab system usb` | Raw block-device write |
| **git** | Future repos features; backup may use git | Protocol and object store |
| **mise** | `lab toolchain` | Multi-language version manager |
| **brew** / **apt** / **dnf** / **rpm-ostree** | `lab pkg` | OS package managers |

### Interim homelab script

| Script | Command | Plan |
|--------|---------|------|
| `tools/gitlab/backup_account.py` | `lab repos backup` | Port to Go (GitLab API + git) |
| `scripts/install-server-deps.sh` | bootstrap `script:` step | Port or keep as one-off server step |

YouTube downloads are intentionally **not** part of this CLI; use homelab scripts or `yt-dlp` directly if needed.
