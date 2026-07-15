// Package tempo provides the tempo compose service.
package tempo

import "github.com/bartrosa/homelab-cli/internal/services"

func init() {
	services.Register(services.NewManagedService(
		services.ServiceMeta{
			ID: "tempo", DisplayName: "Tempo", Category: services.CategoryObservability,
			Description: "Distributed tracing backend",
		},
		services.ConfigSchema{Fields: []services.Field{
			services.PortField("port", "HTTP port", 3200),
		}},
		nil, nil,
	))
}
