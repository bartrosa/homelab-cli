package services

// DefaultTemplateData builds the standard template context for a service.
func DefaultTemplateData(id string, values map[string]any, dirs map[string]string) (map[string]any, error) {
	return defaultTemplateData(id, values, dirs)
}
