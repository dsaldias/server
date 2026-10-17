package xlogin

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

var (
	ADMIN_HOME   = "/adminx/"
	ADMIN_LOGIN  = "/adminx/login"
	ADMIN_INICIO = "/adminx/inicio"
)

func RutasAdmin(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{DB: db}
	rutas := []*utilidad.Handlers2{}

	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_HOME, H: http.HandlerFunc(controller.LoginRender)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_LOGIN, H: http.HandlerFunc(controller.LoginSubmit)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_INICIO, H: http.HandlerFunc(controller.InicioRender)})

	return rutas
}
