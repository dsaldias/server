package dataauth

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/dsaldias/server/dataauth/utils"
	"github.com/dsaldias/server/dataauth/xnotificaciones"
	"github.com/dsaldias/server/graph_auth"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/vektah/gqlparser/v2/ast"
)

func Iniciar(srv *handler.Server, schema *graphql.ExecutableSchema, db *sql.DB) {

	port := os.Getenv("PORT")
	rate := os.Getenv("RATE_LIMIT")

	router := chi.NewRouter()
	router.Use(middleware.Compress(5, "application/json", "application/graphql+json"))
	router.Use(cors.New(cors.Options{
		AllowedOrigins:   utils.GetAllowedOrigins(),
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)
	router.Use(utils.MiddlewareCookie)
	router.Use(utils.AuthMiddleware(db))
	if rate == "1" {
		rl := utils.NewRateLimiter(18, time.Second)
		router.Use(rl.RateMiddleware)
	}

	xnotificaciones.InitializeGlobal()

	resolver2 := &graph_auth.Resolver{DB: db}
	schema2 := graph_auth.NewExecutableSchema(graph_auth.Config{Resolvers: resolver2})
	srv2 := handler.New(schema2)

	srv2.AddTransport(transport.Options{})
	srv2.AddTransport(transport.GET{})
	srv2.AddTransport(transport.POST{})
	srv2.AddTransport(&transport.Websocket{
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

	srv2.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv2.Use(extension.Introspection{})
	srv2.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	show_playground := os.Getenv("PLAYGROUND")
	if show_playground == "1" {
		router.Handle("/auth", playground.Handler("GraphQL Auth", "/query_auth"))
		router.Handle("/app", playground.Handler("GraphQL App", "/query"))
	}

	router.Handle("/query_auth", srv2)
	router.Handle("/ws", srv2)
	router.Handle("/query", srv)
	router.Handle("/ws_app", srv)
	router.Get("/sse", xnotificaciones.SSEHandler)
	router.Handle("/res/*", http.StripPrefix("/res/", http.FileServer(http.Dir("res"))))
	// rest to graphql
	if schema != nil {
		router.Post("/rest/mutation/{operationName}", utils.RestToGraphQlHandler(*schema))
		router.Get("/rest/query/{operationName}", utils.RestToGraphQlHandler(*schema))
	}

	router.Post("/rest_auth/mutation/{operationName}", utils.RestToGraphQlHandler(schema2))
	router.Get("/rest_auth/query/{operationName}", utils.RestToGraphQlHandler(schema2))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
