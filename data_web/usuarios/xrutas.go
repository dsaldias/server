package usuarios

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

var (
	ADMIN_USUARIOS = "/adminx/usuarios"
)

func Rutas(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{DB: db}
	rutas := []*utilidad.Handlers2{}

	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_USUARIOS, H: http.HandlerFunc(controller.UsuariosRender)})
	return rutas
}
