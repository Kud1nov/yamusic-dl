// Package utils предоставляет вспомогательные функции.
package utils

import (
	"regexp"
	"strings"
)

// CleanFileName очищает строку от недопустимых символов для имени файла.
// Оставляет только буквы, цифры и некоторые специальные символы.
func CleanFileName(name string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9 _\-]`)
	clean := reg.ReplaceAllString(name, "")
	return strings.TrimSpace(clean)
}
