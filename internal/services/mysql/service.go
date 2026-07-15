// Package mysql provides the mysql compose service.
package mysql

import "github.com/bartrosa/homelab-cli/internal/services"

func init() {
	services.Register(services.NewManagedService(
		services.ServiceMeta{
			ID: "mysql", DisplayName: "MySQL", Category: services.CategoryDatabase,
			Description: "MySQL 8 relational database",
		},
		services.ConfigSchema{Fields: []services.Field{
			services.PortField("port", "Host port", 3306),
			services.UserField("user", "Root user", "root"),
			services.PasswordField("password", "Root password"),
			{Name: "database", Label: "Default database", Type: services.FieldTypeString, Default: "homelab", Required: true},
		}},
		nil, nil,
	))
}
