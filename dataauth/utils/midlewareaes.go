package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func XXMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			fmt.Println(">> DecryptBodyMiddleware - Inicio")
			defer fmt.Println("<< DecryptBodyMiddleware - Fin")

			// 1. Leer el cuerpo cifrado
			encryptedBodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				fmt.Println("Error leyendo el cuerpo cifrado:", err)
				next.ServeHTTP(w, r)
				// http.Error(w, "Error al leer la petición", http.StatusBadRequest) // Error HTTP
				return
			}
			defer r.Body.Close() // ¡Cerrar el body original después de leerlo!

			fmt.Println(">>>>\n", string(encryptedBodyBytes))
			decryptedBodyReader := bytes.NewReader(encryptedBodyBytes)
			r.Body = io.NopCloser(decryptedBodyReader)

			encryptedBodyString := string(encryptedBodyBytes)
			if strings.TrimSpace(encryptedBodyString) == "" {
				fmt.Println("Cuerpo de la petición vacío o solo espacios en blanco, continuando sin descifrar.")
				next.ServeHTTP(w, r) // Continuar si el cuerpo está vacío
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
