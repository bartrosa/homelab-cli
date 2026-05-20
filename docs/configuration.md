# Configuration

## Precedence

1. **CLI flags** — `--log-level`, `--log-format`, `--config`, `--homelab-root`, `--dry-run`, `--no-color`
2. **Environment** — `LAB_*` prefix; nested keys use `_` (e.g. `LAB_SERVER_HOST`, `LAB_HOMELAB_ROOT`)
3. **YAML file** — `--config` or default `~/.config/homelab-cli/config.yaml`
4. **Built-in defaults** — `internal/config.Default()`

If the config file does not exist, defaults apply. Invalid YAML fails at load time.

## Starter file

Copy [`config.example.yaml`](config.example.yaml):

```bash
mkdir -p ~/.config/homelab-cli
cp docs/config.example.yaml ~/.config/homelab-cli/config.yaml
```

## Keys

### Global

| Key | Env (examples) | Description |
|-----|----------------|-------------|
| `log_level` | `LAB_LOG_LEVEL` | `debug`, `info`, `warn`, `error` |
| `log_format` | `LAB_LOG_FORMAT` | `text` or `json` |

### `homelab`

| Key | Env | Description |
|-----|-----|-------------|
| `homelab.root` | `LAB_HOMELAB_ROOT` | Absolute or `~/` path to the personal homelab repository. Required for compose stacks, templates, postgres YAML, and GitLab backup script. Overridable with `--homelab-root`. |

### `server`

Default target for `lab ssh sync`, `lab server run`, and `lab server deploy`.

| Key | Env | Description |
|-----|-----|-------------|
| `server.host` | `LAB_SERVER_HOST` | Hostname or IP |
| `server.user` | `LAB_SERVER_USER` | SSH user (default `root`) |
| `server.port` | `LAB_SERVER_PORT` | SSH port (default `22`) |
| `server.path` | `LAB_SERVER_PATH` | Remote directory containing the homelab checkout |

### `ssh.hosts`

Map of alias → host for `lab ssh connect <alias>`.

```yaml
ssh:
  hosts:
    homelab:
      host: 192.168.1.10
      user: bart
      port: 22
      identity_file: ~/.ssh/id_ed25519
```

### `bootstrap`

| Key | Description |
|-----|-------------|
| `bootstrap.default_profile` | Default profile name |
| `bootstrap.profiles` | Custom profiles (same step types as embedded YAML: `pkg`, `toolchain`, `script`) |

Embedded profiles live in `internal/bootstrap/profiles/*.yaml` and are always available.

### `repos`

| Key | Description |
|-----|-------------|
| `repos.root` | Local directory for cloned repos (future) |
| `repos.backup_dir` | GitLab backup destination |
| `repos.providers[]` | `name`, `kind` (`github`, `gitlab`, …), `host`, `token_env` |

### `services`

| Key | Description |
|-----|-------------|
| `services.runtime` | `podman-compose` or `docker` |
| `services.stacks_dir` | Reserved; stacks are resolved from `homelab.root` today |

### `cluster` / `storage`

Used by future commands; safe to set early.

| Key | Description |
|-----|-------------|
| `cluster.kubeconfig` | Path to kubeconfig |
| `cluster.context` | Default context name |
| `storage.endpoint` | S3-compatible endpoint |
| `storage.access_key` | Access key (prefer env for secrets) |

## Example

```yaml
log_level: info
log_format: text

homelab:
  root: ~/Projects/PERSONAL/homelab

server:
  host: 192.168.1.10
  user: bart
  port: 22
  path: ~/homelab

bootstrap:
  default_profile: laptop-macos
  profiles: {}

repos:
  root: ~/src
  backup_dir: ~/backups/repos/gitlab
  providers:
    - name: gitlab-personal
      kind: gitlab
      host: gitlab.com
      token_env: GITLAB_TOKEN

services:
  runtime: podman-compose

ssh:
  hosts:
    homelab:
      host: 192.168.1.10
      user: bart
      port: 22

cluster:
  kubeconfig: ~/.kube/config
  context: homelab
```

## PostgreSQL apply

`lab postgres apply` reads a separate YAML file (not the main lab config), typically:

`$HOMELAB_ROOT/postgres/config/instances.yaml`

Set `POSTGRES_ADMIN_PASSWORD` or `PGPASSWORD` before apply.
