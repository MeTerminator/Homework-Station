package service

import (
	"strings"
)

func AllowedExtension(name, ext string) bool {
	if strings.TrimSpace(ext) == "" {
		return true
	}
	name = strings.ToLower(name)
	for _, e := range strings.FieldsFunc(ext, func(r rune) bool { return r == ',' || r == ';' || r == ' ' || r == '，' }) {
		e = "." + strings.TrimPrefix(strings.ToLower(strings.TrimSpace(e)), ".")
		if strings.HasSuffix(name, e) {
			return true
		}
	}
	return false
}
