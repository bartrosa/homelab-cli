// Package prompt provides interactive stdin prompting for service init.
package prompt

import (
	"fmt"
)

// Prompter collects user input during service initialization.
type Prompter interface {
	AskString(label, defaultValue string) (string, error)
	AskPassword(label string) (string, error)
	AskBool(label string, defaultValue bool) (bool, error)
	AskSelect(label string, options []string, defaultIndex int) (int, error)
	AskMultiSelect(label string, options []string, defaultIndexes []int) ([]int, error)
}

// AskAll prompts for every field in schema not already present in existing.
func AskAll(p Prompter, schema Schema, existing map[string]any) (map[string]any, error) {
	out := map[string]any{}
	for _, f := range schema.Fields {
		if v, ok := existing[f.Name]; ok && !isEmpty(v) {
			out[f.Name] = v
			continue
		}
		val, err := askField(p, f)
		if err != nil {
			return nil, err
		}
		out[f.Name] = val
	}
	return out, nil
}

// ValidateSchema ensures required fields are present for non-interactive init.
func ValidateSchema(schema Schema, values map[string]any) error {
	for _, f := range schema.Fields {
		if !f.Required {
			continue
		}
		v, ok := values[f.Name]
		if !ok || isEmpty(v) {
			return fmt.Errorf("missing required field %q", f.Name)
		}
	}
	return nil
}

func askField(p Prompter, f Field) (any, error) {
	label := f.Label
	if label == "" {
		label = f.Name
	}
	switch f.Type {
	case FieldTypeString:
		def := stringDefault(f.Default)
		return p.AskString(label, def)
	case FieldTypePassword:
		return p.AskPassword(label)
	case FieldTypeInt:
		def := stringDefault(f.Default)
		s, err := p.AskString(label, def)
		if err != nil {
			return nil, err
		}
		var n int
		if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
			return nil, fmt.Errorf("%s: invalid integer %q", f.Name, s)
		}
		return n, nil
	case FieldTypeBool:
		def := false
		if b, ok := f.Default.(bool); ok {
			def = b
		}
		return p.AskBool(label, def)
	case FieldTypeSelect:
		idx := 0
		if n, ok := f.Default.(int); ok {
			idx = n
		}
		return p.AskSelect(label, f.Options, idx)
	case FieldTypeMultiSelect:
		defs := []int{}
		if ss, ok := f.Default.([]string); ok {
			for _, s := range ss {
				for i, opt := range f.Options {
					if opt == s {
						defs = append(defs, i)
					}
				}
			}
		}
		idxs, err := p.AskMultiSelect(label, f.Options, defs)
		if err != nil {
			return nil, err
		}
		var selected []string
		for _, i := range idxs {
			if i >= 0 && i < len(f.Options) {
				selected = append(selected, f.Options[i])
			}
		}
		return selected, nil
	default:
		def := stringDefault(f.Default)
		return p.AskString(label, def)
	}
}

func stringDefault(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	switch t := v.(type) {
	case string:
		return t == ""
	case []string:
		return len(t) == 0
	case int:
		return false
	case bool:
		return false
	default:
		return false
	}
}

// Field mirrors services schema for prompt without import cycle.
type Field struct {
	Name     string
	Label    string
	Type     FieldType
	Default  any
	Required bool
	Options  []string
}

// FieldType describes a config field kind.
type FieldType string

// Field types supported by the prompt engine.
const (
	FieldTypeString      FieldType = "string"
	FieldTypeInt         FieldType = "int"
	FieldTypeBool        FieldType = "bool"
	FieldTypePassword    FieldType = "password"
	FieldTypeSelect      FieldType = "select"
	FieldTypeMultiSelect FieldType = "multiselect"
)

// Schema is a list of prompt fields.
type Schema struct {
	Fields []Field
}
