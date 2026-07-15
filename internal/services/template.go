package services

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"strings"
	"text/template"
)

//go:embed templates/services/**
var serviceTemplates embed.FS

// Render executes a service template with data.
func Render(serviceID, tmplName string, data any) (string, error) {
	tmplPath := path.Join("templates", "services", serviceID, tmplName)
	raw, err := serviceTemplates.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("read template %s: %w", tmplPath, err)
	}
	tmpl, err := template.New(tmplName).Funcs(templateFuncs()).Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", tmplPath, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %s: %w", tmplPath, err)
	}
	return strings.TrimSpace(buf.String()) + "\n", nil
}

// ListTemplates returns embedded template names for a service.
func ListTemplates(serviceID string) ([]string, error) {
	prefix := path.Join("templates", "services", serviceID)
	var names []string
	err := walkEmbed(serviceTemplates, prefix, func(name string) error {
		base := strings.TrimPrefix(name, prefix+"/")
		if base != "" && !strings.Contains(base, "/") {
			names = append(names, base)
		}
		return nil
	})
	return names, err
}

func walkEmbed(fsys embed.FS, dir string, fn func(string) error) error {
	entries, err := fsys.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		full := path.Join(dir, e.Name())
		if e.IsDir() {
			if err := walkEmbed(fsys, full, fn); err != nil {
				return err
			}
			continue
		}
		if err := fn(full); err != nil {
			return err
		}
	}
	return nil
}
