package utils

import (
	"context"
	"database/sql"

	"github.com/99designs/gqlgen/graphql/handler/transport"
)

type contextKey string

const (
	uidContextKey contextKey = "uid"
)

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

func CtxUserIDWs(ctx context.Context, db *sql.DB, metodo string) string {
	algo := ctx.Value(uidContextKey)
	if algo == nil {
		return "general"
	}

	clains, ok := algo.(*string)
	if ok {
		return *clains
	}

	return "general"
}
