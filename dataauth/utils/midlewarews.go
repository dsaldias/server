package utils

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler/transport"
)

type contextKey string

const (
	uidContextKey contextKey = "uid"
)

func WsMiddlewareCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := CtxGetCookie(r)
		if err == nil {
			ctxx := context.WithValue(r.Context(), uidContextKey, &cookie.UserID)
			next.ServeHTTP(w, r.WithContext(ctxx))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func UaserIDMiddleware(db *sql.DB) transport.WebsocketInitFunc {

	return func(ctx context.Context, initPayload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
		uid, ok3 := initPayload["uid"].(string)
		pay := &transport.InitPayload{}

		if !ok3 {
			return ctx, pay, nil
		}

		ctxx := context.WithValue(ctx, uidContextKey, &uid)

		return ctxx, pay, nil
	}
}

func CtxUserIDWs(ctx context.Context, db *sql.DB, metodo string) *string {
	algo := ctx.Value(uidContextKey)
	if algo == nil {
		return nil
	}

	clains, ok := algo.(*string)
	if ok {
		return clains
	}

	return nil
}
