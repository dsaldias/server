package usuarios

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

var (
	ADMIN_USUARIOS           = "/adminx/usuarios"
	ADMIN_USUARIOS_LISTAHTML = ADMIN_USUARIOS + "/listahtml"
	ADMIN_USUARIOS_FORMHTML  = ADMIN_USUARIOS + "/formulariohtml"
	ADMIN_USUARIOS_VERHTML   = ADMIN_USUARIOS + "/visualizarhtml"
	ADMIN_USUARIOS_EDITHTML  = ADMIN_USUARIOS + "/editarhtml"
	ADMIN_USUARIOS_CREAR     = ADMIN_USUARIOS + "/crear"
	ADMIN_USUARIOS_EDITAR    = ADMIN_USUARIOS + "/editar"
)

func Rutas(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{DB: db}
	rutas := []*utilidad.Handlers2{}

	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_USUARIOS, H: http.HandlerFunc(controller.UsuariosRender)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_USUARIOS_LISTAHTML, H: http.HandlerFunc(controller.UsuariosListaHTML)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_USUARIOS_FORMHTML, H: http.HandlerFunc(controller.UsuarioFormularioRender)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_USUARIOS_VERHTML, H: http.HandlerFunc(controller.UsuarioVisualizarRender)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_USUARIOS_EDITHTML, H: http.HandlerFunc(controller.UsuarioEditarRender)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_USUARIOS_CREAR, H: http.HandlerFunc(controller.UsuarioCrear)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_USUARIOS_EDITAR, H: http.HandlerFunc(controller.UsuarioEditar)})
	return rutas
}
