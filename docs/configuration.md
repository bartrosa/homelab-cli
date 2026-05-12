# Configuration

## Precedence

1. CLI flags (`--log-level`, `--log-format`, `--config`, …)
2. Environment variables with the `LAB_` prefix (nested keys use `_`, e.g. `LAB_SERVICES_RUNTIME`)
3. YAML file (`--config`, default `~/.config/homelab-cli/config.yaml`)
4. Built-in defaults in `internal/config`

## Example `config.yaml`

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
    - name: gitlab-work
      kind: gitlab
      host: gitlab.com
      token_env: GITLAB_TOKEN

services:
  stacks_dir: ~/.config/homelab-cli/stacks
  runtime: podman # docker|podman

cluster:
  kubeconfig: ~/.kube/config
  context: homelab

storage:
  endpoint: https://minio.lan:9000
  access_key: REPLACE_ME
```

Missing files fall back to defaults; malformed YAML is an error.
