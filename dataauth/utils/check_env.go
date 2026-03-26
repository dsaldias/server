package utils

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func VerificarEnv() {
	// Lista de variables base requeridas
	requiredVars := []string{
		"PORT",
		"DB_USER",
		"DB_PASS",
		"DB_HOST",
		"DB_NAME",
		"PERM_EXTERNO",
		"EXTERNAL_AUTH",
		"EXTERNAL_ME",
		"PLAYGROUND",
		"RATE_LIMIT",
		"DECODE_PASS_KEY",
		"TOKEN_DURATION_MIN",
		"AUTH_SHOW_NAME_PERMISO",
		"SEND_NOTI_LOGIN",
		"DEFAULT_UNIDAD_OAUTH",
		"DEFAULT_ROL_OAUTH",
		"DEFAULT_ROL_EXTER",
		"OAUTH_EMAILS_PERM",
		"DB_CONN_LIFETIME_MIN",
		"DB_MAX_OPEN",
		"DB_MAX_IDLE",
		"ALLOWED_ORIGINS",
	}

	envMap, err := godotenv.Read()
	if err != nil {
		fmt.Println("ERROR LEYENDO EL ARCHIVO .env PARA VALIDAR ENTORNO: ", err)
		return
	}

	// Verificar variables faltantes
	var missingVars []string
	for _, key := range requiredVars {
		if _, exists := envMap[key]; !exists {
			missingVars = append(missingVars, key)
		}
	}

	// Resultado
	if len(missingVars) > 0 {
		log.Fatalf("Variables faltantes en el .env: %v", missingVars)
	}
}

func GetAllowedOrigins() []string {
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	allowedOrigins := []string{"*"}
	if allowedOriginsEnv != "" {
		parts := strings.Split(allowedOriginsEnv, ",")
		allowedOrigins = make([]string, 0, len(parts))
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				allowedOrigins = append(allowedOrigins, s)
			}
		}
	}
	return allowedOrigins
}
