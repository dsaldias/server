package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed skills
var skillsFS embed.FS

//go:embed docs/CLAUDE.md.tmpl
var claudeMDTemplate string

//go:embed sql/database-app.sql.tmpl
var sqlAppTemplate string

//go:embed taskfile/Taskfile.yml
var taskFilex string

//go:embed taskfile/components.json
var componetesjsonx string

//go:embed taskfile/globals.css
var inputCcs string

//go:embed sql/database.sql
var sqlRBAC string

var consumerTools = []string{
	"github.com/a-h/templ/cmd/templ@v0.3.1070",
	"github.com/axadrn/shadcn-templ/v2/cmd/shadcn-templ@v2.0.0-beta.15",
}

// Dependencias runtime que usan los paquetes frontend de la librería.
// Se registran en el consumidor porque no son herramientas de Go y deben
// quedar disponibles en su go.mod/go.sum antes de compilar serverx.go.
var consumerDependencies = []string{
	"github.com/starfederation/datastar-go@v1.2.2",
}

func Init() {
	module := getModuleName()

	contentx := fmt.Sprintf(`
package main

import (
	"%s/app"
	"%s/app/front"
	"%s/graph" 
	"time"
	"uuid"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/coder/websocket"
	"github.com/dsaldias/server/data_web/xutility"
	"github.com/dsaldias/server/dataauth"
	"github.com/dsaldias/server/dataauth/utilidad"
	"github.com/vektah/gqlparser/v2/ast"
)

// # go run github.com/99designs/gqlgen generate
// # go run github.com/99designs/gqlgen generate --config gqlgen_auth.yml
// # CGO_ENABLED=0 go build -ldflags="-s -w" -o appname serverx.go
// # scp appname root@185.203.216.16:/root/apps/appname/
func main() {

	db := utilidad.Conexion()
	resolver := graph.Resolver{DB: db}
	schema := graph.NewExecutableSchema(graph.Config{Resolvers: &resolver})
	srv := handler.New(schema)

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(&transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		PingPongInterval:      10 * time.Second,
		Implementation: transport.CoderWebsocketImplementation{
			AcceptOptions: websocket.AcceptOptions{
				OriginPatterns: []string{"*"},
			},
		},
		InitFunc: utilidad.UaserIDMiddleware(db),
	})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	app.LoadCustomEvents()

	ssr := front.RutasFront(db)

	// para dev, actualizamos los assets
	xutility.AssetVersion = uuid.New().String()

	dataauth.Iniciar(srv, &schema, db, ssr, nil)
}

`, module, module, module)

	contentenv := fmt.Sprintf(`
PORT=8038
DB_USER=root
DB_PASS=S1nclave
DB_HOST=127.0.0.1
DB_NAME=%s
# EXTERNO
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

	"github.com/dsaldias/server/dataauth/utilidad"
)

/*
Aqui puedes setear sus funciones cuando ocurre alguna accion
por ejemplo cuando se registra un usuario externo
defina su escuchador
*/
func LoadCustomEvents() {
	// EDIT THIS FILE IN YOUR APP
	utilidad.SetOnUserExternalCreate(func(db *sql.DB, id, u, p string) {})
	utilidad.SetOnUserRelogin(func(db *sql.DB, id, u, p string) {})
	utilidad.SetOnTicketCreated(func(db *sql.DB, id string) {})
	utilidad.SetOnUserExternalLogin(func(db *sql.DB, username, pass string) {})
}

	`

	contentresolver := `package graph

import "database/sql"

type Resolver struct {
	DB *sql.DB
}
`

	content_rutas_cli := `
package front

import (
	"database/sql"
	"embed"
	"net/http"

	"github.com/dsaldias/server/data_web/xlogin"
	"github.com/dsaldias/server/dataauth/utilidad"
)

//go:embed asset/*
var Asset embed.FS

/*
// PARA UNA RUTA PUBLICA MAPEA A UN COMPONENTE QUE CARGUE LO SIGUIENTE:
package publico
*/

var (
	WEB_PATH_BASE = "/webc/"
)

func RutasFront(db *sql.DB) []*utilidad.Handlers2 {
	ssr := []*utilidad.Handlers2{}
	conf := xlogin.Default()
	rutitas := xlogin.RutasAdmin(db, conf)
	ssr = append(ssr, rutitas...)

	tcontroller := TestController{DB: db}

	ssr = append(ssr, &utilidad.Handlers2{Path: WEB_PATH_BASE, H: http.HandlerFunc(tcontroller.Listar)})

	return ssr
}

`

	content_controller1 := `
package front

import (
	"database/sql"
	"net/http"
)

type TestController struct {
	DB *sql.DB 
}

func (c *TestController) Listar(w http.ResponseWriter, r *http.Request) {

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

	if err := os.MkdirAll("app/front", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ error creando directorio app/front/: %v\n", err)
	} else {
		escribirArchivo("app/front/rutas_front.go", []byte(content_rutas_cli))
	}

	if err := os.MkdirAll("app/front/asset", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ error creando directorio app/front/asset/: %v\n", err)
	} else {
		escribirArchivo("app/front/asset/.kepp", []byte(""))
	}

	if err := os.MkdirAll("app/front/asset/css", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ error creando directorio app/front/asset/: %v\n", err)
	} else {
		escribirArchivo("app/front/asset/css/globals.css", []byte(inputCcs))
	}

	if err := os.MkdirAll("app/front/utils", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ error creando directorio app/front/utils/: %v\n", err)
	}

	if err := os.MkdirAll("app/front/componentes", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ error creando directorio app/front/componentes/: %v\n", err)
	} else {
		escribirArchivo("app/front/test.go", []byte(content_controller1))
	}

	if err := os.MkdirAll("app/back", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ error creando directorio app/back/: %v\n", err)
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
	instalarDependenciasConsumidor()
	instalarHerramientasConsumidor()
	ejecutarGoModTidy()
}

func instalarDependenciasConsumidor() {
	for _, dependency := range consumerDependencies {
		cmd := exec.Command("go", "get", dependency)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  no fue posible registrar la dependencia %s: %v\n", dependency, err)
			fmt.Fprintln(os.Stderr, "   puedes registrarla manualmente con: go get "+dependency)
			continue
		}
		fmt.Println("✅ dependencia registrada: " + dependency)
	}
}

func instalarHerramientasConsumidor() {
	for _, tool := range consumerTools {
		cmd := exec.Command("go", "get", "-tool", tool)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  no fue posible registrar la herramienta %s: %v\n", tool, err)
			fmt.Fprintln(os.Stderr, "   puedes registrarla manualmente con: go get -tool "+tool)
			continue
		}
		fmt.Println("✅ herramienta registrada: " + tool)
	}
}

func ejecutarGoModTidy() {
	if _, err := os.Stat("go.mod"); err != nil {
		fmt.Println("ℹ️  no se encontró go.mod; omitiendo go mod tidy")
		return
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  no fue posible completar go mod tidy automáticamente: %v\n", err)
		fmt.Fprintln(os.Stderr, "   ejecuta nuevamente: go mod tidy")
		return
	}
	fmt.Println("✅ dependencias del consumidor sincronizadas con go mod tidy")
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
	sqlDir := filepath.Join("assets", "sqls")
	dest := filepath.Join(sqlDir, "database-"+shortName+".sql")
	dest2 := "Taskfile.yml"
	dest3 := "components.json"

	if err := os.MkdirAll(sqlDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ error creando directorio %s/: %v\n", sqlDir, err)
		return
	}

	escribirArchivo(filepath.Join(sqlDir, "database.sql"), []byte(sqlRBAC))
	content := strings.ReplaceAll(sqlAppTemplate, "{{MODULE}}", module)
	escribirArchivo(dest, []byte(content))

	content2 := strings.ReplaceAll(taskFilex, "{{MODULE}}", module)
	escribirArchivo(dest2, []byte(content2))

	content3 := strings.ReplaceAll(componetesjsonx, "{{MODULE}}", module)
	escribirArchivo(dest3, []byte(content3))
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
