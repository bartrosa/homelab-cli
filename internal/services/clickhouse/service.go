// Package clickhouse provides the clickhouse compose service.
package clickhouse

import "github.com/bartrosa/homelab-cli/internal/services"

func init() {
	services.Register(services.NewManagedService(
		services.ServiceMeta{
			ID: "clickhouse", DisplayName: "ClickHouse", Category: services.CategoryDatabase,
			Description: "Column-oriented OLAP database",
		},
		services.ConfigSchema{Fields: []services.Field{
			services.PortField("port", "HTTP port", 8123),
			services.UserField("user", "Default user", "default"),
			services.PasswordField("password", "Default password"),
		}},
		nil, nil,
	))
}
