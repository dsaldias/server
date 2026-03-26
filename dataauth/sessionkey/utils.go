package sessionkey

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"

	"github.com/dsaldias/server/graph_auth/model"
)

func parse(r *sql.Row, t *model.SessionKey) error {
	return r.Scan(
		&t.ID,
		&t.UsuarioID,
		&t.Key,
		&t.Apikey,
		&t.Expire,
		&t.FechaRegistro,
		&t.UserEstado,
	)
}

func generarStringUnico() (string, error) {
	// Define el tamaño de los bytes aleatorios
	numBytes := 32 // Ajusta este valor según el tamaño deseado después de codificar
	randomBytes := make([]byte, numBytes)

	// Genera los bytes aleatorios
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Codifica los bytes en Base64 y corta a 50 caracteres
	uniqueString := base64.RawURLEncoding.EncodeToString(randomBytes)
	if len(uniqueString) > 50 {
		uniqueString = uniqueString[:50]
	}

	return uniqueString, nil
}
