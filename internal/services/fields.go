package services

// PortField is a required integer port prompt field.
func PortField(name, label string, def int) Field {
	return Field{
		Name: name, Label: label, Type: FieldTypeInt, Default: def, Required: true,
	}
}

// PasswordField is a required sensitive password prompt field.
func PasswordField(name, label string) Field {
	return Field{
		Name: name, Label: label, Type: FieldTypePassword, Required: true, Sensitive: true,
	}
}

// UserField is a required string user prompt field.
func UserField(name, label, def string) Field {
	return Field{
		Name: name, Label: label, Type: FieldTypeString, Default: def, Required: true,
	}
}
