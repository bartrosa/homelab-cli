// Package weaviate provides the weaviate compose service.
package weaviate

import "github.com/bartrosa/homelab-cli/internal/services"

func init() {
	services.Register(services.NewManagedService(
		services.ServiceMeta{
			ID: "weaviate", DisplayName: "Weaviate", Category: services.CategoryVector,
			Description: "Vector database with GraphQL API",
		},
		services.ConfigSchema{Fields: []services.Field{
			services.PortField("port", "HTTP port", 8080),
			services.PasswordField("api_key", "API key"),
		}},
		nil, nil,
	))
}
