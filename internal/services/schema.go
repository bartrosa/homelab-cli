package services

// FieldType describes a config field kind.
type FieldType string

// Field types supported by service config schemas.
const (
	FieldTypeString      FieldType = "string"
	FieldTypeInt         FieldType = "int"
	FieldTypeBool        FieldType = "bool"
	FieldTypePassword    FieldType = "password"
	FieldTypeSelect      FieldType = "select"
	FieldTypeMultiSelect FieldType = "multiselect"
)

// Field is one user-configurable service setting.
type Field struct {
	Name        string
	Label       string
	Type        FieldType
	Default     any
	Required    bool
	Options     []string
	Sensitive   bool
	Description string
}

// ConfigSchema describes interactive init fields for a service.
type ConfigSchema struct {
	Fields []Field
}

// FieldByName returns a field definition or nil.
func (s ConfigSchema) FieldByName(name string) *Field {
	for i := range s.Fields {
		if s.Fields[i].Name == name {
			return &s.Fields[i]
		}
	}
	return nil
}
