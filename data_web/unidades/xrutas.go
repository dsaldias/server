package unidades

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

const ADMIN_UNIDADES = "/adminx/unidades"

func Rutas(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{}
	return []*utilidad.Handlers2{{Path: ADMIN_UNIDADES, H: http.HandlerFunc(controller.UnidadesRender)}}
}
