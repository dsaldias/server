package main

import (
	"bufio"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

//go:embed sql/database.sql
var sqlFS embed.FS

func SetupDB() {
	env, err := leerEnv(".env")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ no se pudo leer .env: %v\n", err)
		os.Exit(1)
	}

	host := envVal(env, "DB_HOST", "127.0.0.1")
	user := envVal(env, "DB_USER", "")
	pass := envVal(env, "DB_PASS", "")
	name := envVal(env, "DB_NAME", "")

	if user == "" || name == "" {
		fmt.Fprintln(os.Stderr, "❌ DB_USER y DB_NAME son requeridos en .env")
		os.Exit(1)
	}

	// Primero conectar sin base de datos para crearla si no existe
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/", user, pass, host)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ error conectando a MySQL: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ no se pudo conectar a MySQL (%s): %v\n", host, err)
		os.Exit(1)
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", name))
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ error creando base de datos: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ base de datos `%s` lista\n", name)

	// Reconectar con la base de datos y multiStatements habilitado
	dsnDB := fmt.Sprintf("%s:%s@tcp(%s)/%s?multiStatements=true", user, pass, host, name)
	dbConn, err := sql.Open("mysql", dsnDB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ error conectando a `%s`: %v\n", name, err)
		os.Exit(1)
	}
	defer dbConn.Close()

	script, err := sqlFS.ReadFile("sql/database.sql")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ error leyendo script SQL: %v\n", err)
		os.Exit(1)
	}

	if _, err := dbConn.Exec(string(script)); err != nil {
		fmt.Fprintf(os.Stderr, "❌ error ejecutando script SQL: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ tablas RBAC creadas correctamente")
	fmt.Printf("\n   Playground disponible en http://localhost:%s/auth (una vez levantado el servidor)\n", envVal(env, "PORT", "8038"))
}

func leerEnv(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		result[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}
	return result, scanner.Err()
}

func envVal(env map[string]string, key, fallback string) string {
	if v, ok := env[key]; ok && v != "" {
		return v
	}
	return fallback
}
