package xlogin

import (
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

var (
	ADMIN_HOME = "/adminx/"
)

func RutasAdmin() []*utilidad.Handlers2 {
	controller := Controller{}
	rutas := []*utilidad.Handlers2{}

	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_HOME, H: http.HandlerFunc(controller.LoginRender)})

	return rutas
}
