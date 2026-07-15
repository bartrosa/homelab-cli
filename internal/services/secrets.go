package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// RandomPassword returns a URL-safe random password.
func RandomPassword(n int) (string, error) {
	if n <= 0 {
		n = 24
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("random password: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf)[:n], nil
}

// WriteEnvFile writes key=value lines and chmod 600.
func WriteEnvFile(path string, vars map[string]string) error {
	var b strings.Builder
	for k, v := range vars {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(v)
		b.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

// MaskSensitive replaces sensitive substrings for display.
func MaskSensitive(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
}

// FillSecrets generates random values for empty password fields.
func FillSecrets(schema ConfigSchema, values map[string]any) error {
	for _, f := range schema.Fields {
		if f.Type != FieldTypePassword && !f.Sensitive {
			continue
		}
		cur, ok := values[f.Name]
		if ok {
			if s, ok := cur.(string); ok && s != "" {
				continue
			}
		}
		pw, err := RandomPassword(24)
		if err != nil {
			return err
		}
		values[f.Name] = pw
	}
	return nil
}
