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
| `graphrag` | arcadedb, qdrant, minio, postgres |
| `graph-lab` | arcadedb, nebulagraph |

Override or extend via `services.presets` in config YAML.

## Graph databases

### arcadedb

Apache 2.0 multi-model database (graph, document, KV, vector, time-series). Single container with Studio UI, Cypher/GSQL/Gremlin/SQL, optional MongoDB/Redis protocol plugins, optional MCP plugin (gated until verified in upstream image).

```bash
lab services init arcadedb --set databases=knowledge_graph,rag_docs --yes
lab services up arcadedb
lab services connect arcadedb
lab services connect arcadedb --interactive   # bin/console.sh
```

Endpoints: Studio `http://127.0.0.1:2480`, binary `:2424`, HTTP API `/api/v1`.

### nebulagraph

Apache 2.0 distributed graph (CNCF Database Landscape). Four-container MVP stack: `metad0`, `storaged0`, `graphd`, `studio`. openCypher-compatible nGQL. Post-init script auto-registers storage and rotates root password from default `nebula`.

```bash
lab services init nebulagraph --yes
lab services up nebulagraph    # runs post-init after healthy
lab services connect nebulagraph --interactive
```

Endpoints: nGQL `:9669`, Studio `http://127.0.0.1:7001`.

`graph-lab` preset runs both graph databases in parallel (no port collision: ArcadeDB 2480/2424, NebulaGraph 9669/7001). Expect higher resource use (NebulaGraph = 4 containers + JVM for ArcadeDB).

### Graph database licensing decisions

| Candidate | License | Verdict |
|-----------|---------|---------|
| **ArcadeDB** | Apache 2.0 (explicit “forever” commitment) | ✅ Included |
| **NebulaGraph** | Apache 2.0 (CNCF Database Landscape) | ✅ Included |
| Neo4j Community | GPLv3 + proprietary Enterprise | ❌ Copyleft + enterprise lock |
| FalkorDB | SSPL v1 (not OSI-approved) | ❌ SaaS-clause risk |
| Memgraph | BSL 1.1 (not OSI-approved) | ❌ Commercial restrictions |
| ArangoDB | BSL 1.1 since 2024 | ❌ License change |
| JanusGraph | Apache 2.0 | ⏭ Deferred — requires Cassandra/HBase |
| HugeGraph | Apache 2.0 (ASF) | ⏭ Deferred — no Cypher, multi-component |
| Kuzu | MIT (archived Oct 2025, Apple acquisition) | ⏭ Docs only — see embedded alternatives |

Upstream LICENSE references: [ArcadeDB](https://github.com/ArcadeData/arcadedb/blob/main/LICENSE), [NebulaGraph](https://github.com/vesoft-inc/nebula/blob/master/LICENSE).

### Embedded graph alternatives

Unlike embedded relational DBs in `lab stack` (SQLite, DuckDB), embedded graph engines are **not** in the homelab-cli registry. The landscape is unstable: **Kuzu** (MIT) was archived in October 2025 after Apple’s acquisition. Community forks (**LadybugDB**, **bighorn**) are too early for official support.

For local GraphRAG in Python:

- Pin last Kuzu release: `uv pip install kuzu==0.11.3` (MIT)
- Watch LadybugDB / bighorn on GitHub
- Or run `lab services up arcadedb` and connect from your project over HTTP/binary

Embedded graph in `lab stack` may arrive in a future PR when a fork stabilizes its release cycle.

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
| arcadedb | `http://127.0.0.1:2480` (Studio), binary `:2424` |
| nebulagraph | nGQL `:9669`, Studio `http://127.0.0.1:7001` |

Use `lab services connect <id> --interactive` where a CLI client is available inside the container.
