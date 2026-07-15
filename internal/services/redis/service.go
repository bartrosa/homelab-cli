// Package redis provides the redis compose service.
package redis

import "github.com/bartrosa/homelab-cli/internal/services"

func init() {
	services.Register(services.NewManagedService(
		services.ServiceMeta{
			ID: "redis", DisplayName: "Redis", Category: services.CategoryCache,
			Description: "In-memory key-value store",
		},
		services.ConfigSchema{Fields: []services.Field{
			services.PortField("port", "Host port", 6379),
			services.PasswordField("password", "Redis password"),
		}},
		nil, nil,
	))
}
