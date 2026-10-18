package usuarios

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

var (
	ADMIN_USUARIOS            = "/adminx/usuarios"
	ADMIN_USUARIOS_FORMULARIO = ADMIN_USUARIOS + "/formulario"
	ADMIN_USUARIOS_VISUALIZAR = ADMIN_USUARIOS + "/visualizar"
	ADMIN_USUARIOS_EDITAR     = ADMIN_USUARIOS + "/editar"
)

func Rutas(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{DB: db}
	rutas := []*utilidad.Handlers2{}

	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_USUARIOS, H: http.HandlerFunc(controller.UsuariosRender)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_USUARIOS_FORMULARIO, H: http.HandlerFunc(controller.UsuarioFormularioRender)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_USUARIOS_VISUALIZAR, H: http.HandlerFunc(controller.UsuarioVisualizarRender)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_USUARIOS_EDITAR, H: http.HandlerFunc(controller.UsuarioEditarRender)})
	return rutas
}
