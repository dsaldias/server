package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// CN combina las clases base de los componentes Shadcn con las clases
// opcionales definidas por la pantalla.
func CN(classes ...any) string {
	parts := make([]string, 0, len(classes))
	for _, value := range classes {
		if text, ok := value.(string); ok && text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

// RandomID genera identificadores para componentes que necesitan relacionar
// atributos ARIA o estados internos del componente.
func RandomID() string {
	var value [8]byte
	if _, err := rand.Read(value[:]); err == nil {
		return "shadcn-" + hex.EncodeToString(value[:])
	}
	return "shadcn-component"
}
