package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/a-h/templ"
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

func MergeAttributes(attrs ...templ.Attributes) templ.Attributes {
	merged := templ.Attributes{}
	for _, current := range attrs {
		for key, value := range current {
			merged[key] = value
		}
	}
	return merged
}

func DirectionAttributes(ctx context.Context) templ.Attributes {
	if direction, ok := ctx.Value(directionKey{}).(string); ok && direction != "" {
		return templ.Attributes{"data-templ-direction": direction}
	}
	return templ.Attributes{}
}

type directionKey struct{}

func IfElse[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
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
