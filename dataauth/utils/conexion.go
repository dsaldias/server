package utils

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func Conexion() *sql.DB {
	VerificarEnv()
	godotenv.Load()

	dbuser := os.Getenv("DB_USER")
	dbpass := os.Getenv("DB_PASS")
	dbhost := os.Getenv("DB_HOST")
	dbname := os.Getenv("DB_NAME")
	// loc := "America%2FLa_Paz"
	// dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&loc=%s", dbuser, dbpass, dbhost, dbname, loc)
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", dbuser, dbpass, dbhost, dbname)
	db, err := sql.Open("mysql", dsn)
	fmt.Println("dsn: ", dsn)

	if err != nil {
		panic(err)
	}

	er := db.Ping()
	if er != nil {
		panic(er)
	}

	db.SetConnMaxLifetime(time.Minute * time.Duration(getEnvInt("DB_CONN_LIFETIME_MIN", 5)))
	db.SetMaxOpenConns(getEnvInt("DB_MAX_OPEN", 20))
	db.SetMaxIdleConns(getEnvInt("DB_MAX_IDLE", 5))
	return db
}

func getEnvInt(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	n, err := strconv.Atoi(val)
	if err != nil || n < 0 {
		return def
	}
	return n
}
