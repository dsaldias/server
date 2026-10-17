package xlogin

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

var (
	ADMIN_HOME = "/adminx/"
)

func RutasAdmin(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{DB: db}
	rutas := []*utilidad.Handlers2{}

	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_HOME, H: http.HandlerFunc(controller.LoginRender)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_HOME + "login", H: http.HandlerFunc(controller.LoginSubmit)})

	return rutas
}
