// Package register side-effect imports all bundled homelab services.
package register

import (
	_ "github.com/bartrosa/homelab-cli/internal/services/arcadedb"    // register arcadedb service
	_ "github.com/bartrosa/homelab-cli/internal/services/clickhouse"  // register clickhouse service
	_ "github.com/bartrosa/homelab-cli/internal/services/grafana"     // register grafana service
	_ "github.com/bartrosa/homelab-cli/internal/services/loki"        // register loki service
	_ "github.com/bartrosa/homelab-cli/internal/services/minio"       // register minio service
	_ "github.com/bartrosa/homelab-cli/internal/services/mongodb"     // register mongodb service
	_ "github.com/bartrosa/homelab-cli/internal/services/mysql"       // register mysql service
	_ "github.com/bartrosa/homelab-cli/internal/services/nats"        // register nats service
	_ "github.com/bartrosa/homelab-cli/internal/services/nebulagraph" // register nebulagraph service
	_ "github.com/bartrosa/homelab-cli/internal/services/postgres"    // register postgres service
	_ "github.com/bartrosa/homelab-cli/internal/services/prometheus"  // register prometheus service
	_ "github.com/bartrosa/homelab-cli/internal/services/qdrant"      // register qdrant service
	_ "github.com/bartrosa/homelab-cli/internal/services/rabbitmq"    // register rabbitmq service
	_ "github.com/bartrosa/homelab-cli/internal/services/redis"       // register redis service
	_ "github.com/bartrosa/homelab-cli/internal/services/tempo"       // register tempo service
	_ "github.com/bartrosa/homelab-cli/internal/services/valkey"      // register valkey service
	_ "github.com/bartrosa/homelab-cli/internal/services/weaviate"    // register weaviate service
)
