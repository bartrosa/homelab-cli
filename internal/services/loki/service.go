// Package loki provides the loki compose service.
package loki

import "github.com/bartrosa/homelab-cli/internal/services"

func init() {
	services.Register(services.NewManagedService(
		services.ServiceMeta{
			ID: "loki", DisplayName: "Loki", Category: services.CategoryObservability,
			Description: "Log aggregation system",
		},
		services.ConfigSchema{Fields: []services.Field{
			services.PortField("port", "HTTP port", 3100),
		}},
		nil, nil,
	))
}
