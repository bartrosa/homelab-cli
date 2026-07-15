package services

import (
	"strings"
	"text/template"
)

// templateFuncs returns default template helpers (no sprig).
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"default": func(def, val any) any {
			if val == nil {
				return def
			}
			switch v := val.(type) {
			case string:
				if v == "" {
					return def
				}
			case []string:
				if len(v) == 0 {
					return def
				}
			}
			return val
		},
		"has": func(needle string, hay any) bool {
			switch h := hay.(type) {
			case []string:
				for _, item := range h {
					if item == needle {
						return true
					}
				}
			case string:
				return h == needle
			}
			return false
		},
		"join": func(sep string, items any) string {
			switch v := items.(type) {
			case []string:
				return strings.Join(v, sep)
			case []any:
				var ss []string
				for _, item := range v {
					ss = append(ss, toString(item))
				}
				return strings.Join(ss, sep)
			default:
				return toString(items)
			}
		},
	}
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
