package utilidad

import (
	"net/http"
)

type Handlers2 struct {
	Path string
	H    http.Handler
}
