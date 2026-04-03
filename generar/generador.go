package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed skills
var skillsFS embed.FS

//go:embed docs/CLAUDE.md.tmpl
var claudeMDTemplate string

//go:embed sql/database-app.sql.tmpl
var sqlAppTemplate string

//go:embed sql/database.sql
var sqlRBAC string

func Init() {
	module := getModuleName()

	contentx := fmt.Sprintf(`
package main

import (
	"%s/app"
	"%s/graph"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/dsaldias/server/dataauth"
	"github.com/dsaldias/server/dataauth/utils"
)

// # go run github.com/99designs/gqlgen generate
// # go run github.com/99designs/gqlgen generate --config gqlgen_auth.yml
// # CGO_ENABLED=0 go build -ldflags="-s -w" -o appname server.go
// # scp appname root@185.203.216.16:/root/apps/appname/
func main() {

	db := utils.Conexion()
	resolver := graph.Resolver{DB: db}
	schema := graph.NewExecutableSchema(graph.Config{Resolvers: &resolver})
	srv := handler.New(schema)

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(&transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		PingPongInterval:      10 * time.Second,
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		InitFunc: utils.UaserIDMiddleware(db),
	})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	app.LoadCustomEvents()
	dataauth.Iniciar(srv, &schema, db)
}

`, module, module)

	contentenv := fmt.Sprintf(`
PORT=8038
DB_USER=root
DB_PASS=S1nclave
DB_HOST=127.0.0.1
DB_NAME=%s
# EXTERNO
PERM_EXTERNO=0
EXTERNAL_AUTH=
EXTERNAL_ME=
# 

PLAYGROUND=1
RATE_LIMIT=1
DECODE_PASS_KEY=Lf5puh9aSuWEmh9Hx1ctoGSn8Qb5kYnn5lM+RBi7e3c=
TOKEN_DURATION_MIN=60
AUTH_SHOW_NAME_PERMISO=1
SEND_NOTI_LOGIN=1
DEFAULT_UNIDAD_OAUTH=1
DEFAULT_ROL_OAUTH=2
DEFAULT_ROL_EXTER=3
OAUTH_EMAILS_PERM=
DB_CONN_LIFETIME_MIN=5
DB_MAX_OPEN=20
DB_MAX_IDLE=5
ALLOWED_ORIGINS=http://localhost:9200,https://sladia.site,https://esam.edu.bo

`, module)

	contentone := `
package app

import (
	"database/sql"

	"github.com/dsaldias/server/dataauth/utils"
)

/*
Aqui puedes setear sus funciones cuando ocurre alguna accion
por ejemplo cuando se registra un usuario externo
defina su escuchador
*/
func LoadCustomEvents() {
	// EDIT THIS FILE IN YOUR APP
	utils.SetOnUserExternalCreate(func(db *sql.DB, id, u, p string) {})
	utils.SetOnUserRelogin(func(db *sql.DB, id, u, p string) {})
	utils.SetOnTicketCreated(func(db *sql.DB, id string) {})
}

	`

	contentresolver := `package graph

import "database/sql"

type Resolver struct {
	DB *sql.DB
}
`

	escribirArchivo("serverx.go", []byte(contentx))
	escribirArchivo(".env", []byte(contentenv))

	if err := os.MkdirAll("graph", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ error creando directorio graph/: %v\n", err)
	} else {
		parchearResolver("graph/resolver.go", contentresolver)
	}

	if err := os.MkdirAll("app", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ error creando directorio app/: %v\n", err)
	} else {
		escribirArchivo("app/onevents.go", []byte(contentone))
	}

	if _, err := os.Stat("server.go"); err == nil {
		if err := os.Rename("server.go", "server.txt"); err != nil {
			fmt.Fprintf(os.Stderr, "❌ error renombrando server.go: %v\n", err)
		} else {
			fmt.Println("📦 server.go renombrado a server.txt")
		}
	}

	copiarSkills()
	generarClaudeMD(module)
	generarSQLApp(module)
}

func escribirArchivo(path string, content []byte) {
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("⚠️  %s ya existe, no se sobrescribe\n", path)
		return
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ error creando %s: %v\n", path, err)
		return
	}
	fmt.Printf("✅ %s creado\n", path)
}

func generarClaudeMD(module string) {
	const dest = "CLAUDE.md"
	if _, err := os.Stat(dest); err == nil {
		fmt.Printf("⚠️ %s ya existe, no se sobrescribe\n", dest)
		return
	}
	content := strings.ReplaceAll(claudeMDTemplate, "{{MODULE}}", module)
	if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
		panic(err)
	}
	fmt.Println("📚 CLAUDE.md creado")
}

func parchearResolver(path, contentresolver string) {
	data, err := os.ReadFile(path)
	if err != nil {
		// no existe, crearlo
		escribirArchivo(path, []byte(contentresolver))
		return
	}
	content := string(data)
	if strings.Contains(content, "DB *sql.DB") {
		fmt.Printf("⚠️  %s ya tiene DB *sql.DB, no se modifica\n", path)
		return
	}
	// agregar import si no está
	if !strings.Contains(content, `"database/sql"`) {
		content = strings.Replace(content, "package graph\n", "package graph\n\nimport \"database/sql\"\n", 1)
	}
	// expandir el struct vacío
	content = strings.ReplaceAll(content, "type Resolver struct{}", "type Resolver struct {\n\tDB *sql.DB\n}")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ error parcheando %s: %v\n", path, err)
		return
	}
	fmt.Printf("✅ %s actualizado con DB *sql.DB\n", path)
}

func generarSQLApp(module string) {
	parts := strings.Split(module, "/")
	shortName := parts[len(parts)-1]
	dest := filepath.Join("sqls", "database-"+shortName+".sql")

	if err := os.MkdirAll("sqls", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ error creando directorio sqls/: %v\n", err)
		return
	}

	escribirArchivo(filepath.Join("sqls", "database.sql"), []byte(sqlRBAC))
	content := strings.ReplaceAll(sqlAppTemplate, "{{MODULE}}", module)
	escribirArchivo(dest, []byte(content))
}

func copiarSkills() {
	fs.WalkDir(skillsFS, "skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		destPath := filepath.Join(".claude", path)
		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}
		data, err := skillsFS.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return err
		}
		fmt.Printf("📋 %s creado\n", destPath)
		return nil
	})
}
