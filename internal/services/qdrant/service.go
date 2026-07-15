// Package qdrant provides the qdrant compose service.
package qdrant

import "github.com/bartrosa/homelab-cli/internal/services"

func init() {
	services.Register(services.NewManagedService(
		services.ServiceMeta{
			ID: "qdrant", DisplayName: "Qdrant", Category: services.CategoryVector,
			Description: "Vector similarity search engine",
		},
		services.ConfigSchema{Fields: []services.Field{
			services.PortField("port", "HTTP port", 6333),
			services.PortField("grpc_port", "gRPC port", 6334),
			{Name: "api_key", Label: "API key (optional)", Type: services.FieldTypePassword, Sensitive: true},
		}},
		nil, nil,
	))
}
