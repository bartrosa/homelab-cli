// Package rabbitmq provides the rabbitmq compose service.
package rabbitmq

import "github.com/bartrosa/homelab-cli/internal/services"

func init() {
	services.Register(services.NewManagedService(
		services.ServiceMeta{
			ID: "rabbitmq", DisplayName: "RabbitMQ", Category: services.CategoryMessageQueue,
			Description: "AMQP message broker with management UI",
		},
		services.ConfigSchema{Fields: []services.Field{
			services.PortField("port", "AMQP port", 5672),
			services.PortField("mgmt_port", "Management port", 15672),
			services.UserField("user", "Default user", "homelab"),
			services.PasswordField("password", "Default password"),
		}},
		nil, nil,
	))
}
