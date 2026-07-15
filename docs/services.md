# Services catalog

Local compose-backed services managed by `lab services`. All stacks join external network **`homelab-net`** so Grafana can reach Prometheus/Loki/Tempo by DNS.

Config layout per service:

```
~/.config/homelab-cli/services/<id>/
├── compose.yml
├── .env              # chmod 600
├── config/
└── init/
```

Data: `~/.local/share/homelab/services/<id>/data/`

## Presets

| Preset | Services |
|--------|----------|
| `observability` | prometheus, loki, tempo, grafana |
| `ml-stack` | postgres, qdrant, minio, clickhouse |
| `data-lakehouse` | postgres, clickhouse, minio |
| `microservices` | postgres, redis, rabbitmq |
| `vector-search` | qdrant, weaviate |
| `full-obs` | prometheus, grafana, loki, tempo, minio |

Override or extend via `services.presets` in config YAML.

## Relational

### postgres

PostgreSQL 17 with optional plugins: **pgvector**, **postgis**, **timescaledb** (multiselect).

- Single-plugin images use upstream tags; **2+ plugins** generate a custom Dockerfile and `docker compose build` on `up`.
- Init SQL enables selected extensions.
- `expose`: `local` (127.0.0.1), `lan` (0.0.0.0), `tailscale` (tailscale0 IP).

```bash
lab services init postgres --set plugins=pgvector,postgis --yes
lab services up postgres
lab services connect postgres
lab services connect postgres --interactive   # psql
```

### mysql

MariaDB 11 (default) or MySQL 8.4 via `flavor` select.

## Cache / KV

### redis

Redis 7; optional Redis Stack modules (redisjson, redisearch, …). Persistence: none, RDB, or AOF.

### valkey

Valkey 7 — open-source Redis fork (Apache 2.0). Drop-in compatible for most workloads.

## NoSQL / analytics

### mongodb

MongoDB 8; optional replica set mode.

### clickhouse

ClickHouse 24; HTTP (8123) and native (9000) endpoints.

## Message brokers

### rabbitmq

RabbitMQ 3 with management UI; optional plugins (prometheus, shovel, federation, delayed-message).

### nats

NATS 2 with JetStream; auth: none, token, or user/password.

## Vector search

### qdrant

Qdrant v1.12 — REST, gRPC, dashboard UI.

### weaviate

Weaviate 1.27; optional `text2vec-transformers` sidecar container.

## Observability

### prometheus

Prometheus v2.55; configurable retention and scrape configs.

### grafana

Grafana OSS 11.3; auto-provisions datasources when selected at init (prometheus, loki, tempo, postgres, clickhouse). **Soft-fail**: Grafana starts even if a datasource target is down; restart Grafana after bringing up missing services.

### loki

Loki 3.3 single-binary, filesystem storage.

### tempo

Tempo 2.6 for OTLP traces.

## Object storage

### minio

MinIO S3-compatible API + console; optional auto-created buckets at init.

## Vector search libraries

**Faiss** is a C++/Python library, not a server. It does not belong in system-level installs. Use `uv pip install faiss-cpu` (or `faiss-gpu`) inside your project. `lab` does not manage project-level Python dependencies.

## Runtime

`services.runtime`: **`auto`** (default) tries `docker compose` first, then podman compose. Ubuntu prefers Docker; Fedora Silverblue prefers Podman unless configured.

## Secrets

- Random passwords generated at `init` (32-char alphanumeric).
- Config values `env:VARNAME` read from environment at runtime — use `op run --env-file=.op.env -- lab services up …`.
- Sensitive fields masked as `********` in logs and status output.

## Connection examples

| Service | Connect |
|---------|---------|
| postgres | `postgres://user:pass@127.0.0.1:5432/db` |
| redis | `redis-cli -h 127.0.0.1 -p 6379` |
| mongodb | `mongosh mongodb://user:pass@127.0.0.1:27017/db` |
| clickhouse | `clickhouse-client` via compose exec |
| grafana | `http://127.0.0.1:3000` |
| minio | `http://127.0.0.1:9000` (API), `:9001` (console) |

Use `lab services connect <id> --interactive` where a CLI client is available inside the container.
