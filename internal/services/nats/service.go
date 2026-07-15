// Package nats provides the nats compose service.
package nats

import "github.com/bartrosa/homelab-cli/internal/services"

func init() {
	services.Register(services.NewManagedService(
		services.ServiceMeta{
			ID: "nats", DisplayName: "NATS", Category: services.CategoryMessageQueue,
			Description: "Cloud-native messaging system",
		},
		services.ConfigSchema{Fields: []services.Field{
			services.PortField("port", "Client port", 4222),
			services.PortField("monitor_port", "Monitor port", 8222),
		}},
		nil, nil,
	))
}
