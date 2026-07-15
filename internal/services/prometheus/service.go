// Package prometheus provides the prometheus compose service.
package prometheus

import "github.com/bartrosa/homelab-cli/internal/services"

func init() {
	services.Register(services.NewManagedService(
		services.ServiceMeta{
			ID: "prometheus", DisplayName: "Prometheus", Category: services.CategoryObservability,
			Description: "Metrics collection and alerting",
		},
		services.ConfigSchema{Fields: []services.Field{
			services.PortField("port", "HTTP port", 9090),
		}},
		nil, nil,
	))
}
