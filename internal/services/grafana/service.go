// Package grafana provides the grafana compose service.
package grafana

import "github.com/bartrosa/homelab-cli/internal/services"

func init() {
	services.Register(services.NewManagedService(
		services.ServiceMeta{
			ID: "grafana", DisplayName: "Grafana", Category: services.CategoryObservability,
			Description: "Metrics and logs visualization",
		},
		services.ConfigSchema{Fields: []services.Field{
			services.PortField("port", "HTTP port", 3000),
			services.UserField("user", "Admin user", "admin"),
			services.PasswordField("password", "Admin password"),
		}},
		[]string{"prometheus", "loki", "tempo"},
		nil,
	))
}
