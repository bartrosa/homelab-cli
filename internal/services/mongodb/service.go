// Package mongodb provides the mongodb compose service.
package mongodb

import "github.com/bartrosa/homelab-cli/internal/services"

func init() {
	services.Register(services.NewManagedService(
		services.ServiceMeta{
			ID: "mongodb", DisplayName: "MongoDB", Category: services.CategoryDatabase,
			Description: "Document database",
		},
		services.ConfigSchema{Fields: []services.Field{
			services.PortField("port", "Host port", 27017),
			services.UserField("user", "Root user", "root"),
			services.PasswordField("password", "Root password"),
		}},
		nil, nil,
	))
}
