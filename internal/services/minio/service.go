// Package minio provides the minio compose service.
package minio

import "github.com/bartrosa/homelab-cli/internal/services"

func init() {
	services.Register(services.NewManagedService(
		services.ServiceMeta{
			ID: "minio", DisplayName: "MinIO", Category: services.CategoryStorage,
			Description: "S3-compatible object storage",
		},
		services.ConfigSchema{Fields: []services.Field{
			services.PortField("port", "API port", 9000),
			services.PortField("console_port", "Console port", 9001),
			services.UserField("user", "Root user", "minioadmin"),
			services.PasswordField("password", "Root password"),
		}},
		nil, nil,
	))
}
