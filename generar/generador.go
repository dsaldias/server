package main

import (
	"fmt"
	"os"
)

func Init() {
	module := getModuleName()

	contentx := fmt.Sprintf(`
package main

import (
	"net/http"
	"%s/graph"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/dsaldias/server/dataauth"
	"github.com/dsaldias/server/dataauth/utils"
	"app"
)

func main() {

	db := utils.Conexion()
	schema := graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db}})
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

`, module)

	contentenv := `
PORT=8038
DB_USER=root
DB_PASS=S1nclave
DB_HOST=127.0.0.1
DB_NAME=auth
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

`

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

	file := "serverx.go"

	if _, err := os.Stat(file); err == nil {
		fmt.Printf("⚠️ %s ya existe. \n", file)
		return
	}

	err := os.WriteFile(file, []byte(contentx), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("✅ serverx.go creado correctamente")
	//

	fileenv := ".env"

	if _, err := os.Stat(fileenv); err == nil {
		fmt.Printf("⚠️ %s ya existe. \n", fileenv)
		return
	}

	err = os.WriteFile(fileenv, []byte(contentenv), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("✅ .env creado correctamente")
	//

	/* err = os.MkdirAll("app", 0755)
	if err != nil {
		panic(err)
	} */

	path := "app/onevents.go"

	if _, err := os.Stat(path); err == nil {
		fmt.Println("⚠️ onevents.go ya existe, no se sobrescribe")
		return
	}

	err = os.WriteFile(path, []byte(contentone), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("📄 onevents.go creado")
	//

	oldFile := "server.go"
	backupFile := "server.txt"

	if _, err := os.Stat(oldFile); err == nil {
		err := os.Rename(oldFile, backupFile)
		if err != nil {
			panic(err)
		}
		fmt.Println("📦 server.go renombrado a server.txt")
	}
}
