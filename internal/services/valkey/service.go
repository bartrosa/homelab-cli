// Package valkey provides the valkey compose service.
package valkey

import "github.com/bartrosa/homelab-cli/internal/services"

func init() {
	services.Register(services.NewManagedService(
		services.ServiceMeta{
			ID: "valkey", DisplayName: "Valkey", Category: services.CategoryCache,
			Description: "Redis-compatible in-memory datastore",
		},
		services.ConfigSchema{Fields: []services.Field{
			services.PortField("port", "Host port", 6379),
			services.PasswordField("password", "Valkey password"),
		}},
		nil, nil,
	))
}
