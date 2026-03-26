package main

import (
	"fmt"
	"os"
)

func Init() {
	module := getModuleName()

	content := fmt.Sprintf(`
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

	dataauth.Iniciar(srv, &schema, db)
}

`, module)

	err := os.WriteFile("serverx.go", []byte(content), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("serverx.go creado ✨")
}
