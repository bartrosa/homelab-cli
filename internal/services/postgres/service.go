// Package postgres provides the postgres compose service.
package postgres

import (
	"github.com/bartrosa/homelab-cli/internal/services"
)

func init() {
	svc := services.NewManagedService(
		services.ServiceMeta{
			ID:          "postgres",
			DisplayName: "PostgreSQL",
			Category:    services.CategoryDatabase,
			Description: "Relational database with optional pgvector, postgis, timescaledb plugins",
		},
		services.ConfigSchema{Fields: []services.Field{
			services.PortField("port", "Host port", 5432),
			services.UserField("user", "Superuser name", "postgres"),
			services.PasswordField("password", "Superuser password"),
			{
				Name: "database", Label: "Default database", Type: services.FieldTypeString,
				Default: "homelab", Required: true,
			},
			{
				Name: "plugins", Label: "Extensions", Type: services.FieldTypeMultiSelect,
				Options: []string{"pgvector", "postgis", "timescaledb"},
				Default: []string{},
			},
		}},
		nil,
		map[string]string{"Dockerfile.tmpl": "Dockerfile"},
	)
	svc.TemplateDataFn = templateData
	services.Register(svc)
}

func templateData(values map[string]any, dirs map[string]string) (any, error) {
	data, err := services.DefaultTemplateData("postgres", values, dirs)
	if err != nil {
		return nil, err
	}
	plugins, _ := values["plugins"].([]string)
	data["Plugins"] = plugins
	data["BuildImage"] = len(plugins) > 0
	return data, nil
}
