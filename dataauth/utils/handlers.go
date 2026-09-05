package utils

import (
	"net/http"
)

type Handlers2 struct {
	Path string
	// H    *handler.Server
	H http.Handler
}
