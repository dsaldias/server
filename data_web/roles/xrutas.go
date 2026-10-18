package roles

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

const (
	ADMIN_ROLES           = "/adminx/roles"
	ADMIN_ROLES_LISTAHTML = ADMIN_ROLES + "/listahtml"
	ADMIN_ROLES_FORMHTML  = ADMIN_ROLES + "/formulariohtml"
	ADMIN_ROLES_VERHTML   = ADMIN_ROLES + "/visualizarhtml"
	ADMIN_ROLES_EDITHTML  = ADMIN_ROLES + "/editarhtml"
	ADMIN_ROLES_CREAR     = ADMIN_ROLES + "/crear"
	ADMIN_ROLES_EDITAR    = ADMIN_ROLES + "/editar"
)

func Rutas(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{DB: db}
	return []*utilidad.Handlers2{
		{Path: ADMIN_ROLES, H: http.HandlerFunc(controller.RolesRender)},
		{Path: ADMIN_ROLES_LISTAHTML, H: http.HandlerFunc(controller.RolesListaHTML)},
		{Path: ADMIN_ROLES_FORMHTML, H: http.HandlerFunc(controller.RolFormularioRender)},
		{Path: ADMIN_ROLES_VERHTML, H: http.HandlerFunc(controller.RolVisualizarRender)},
		{Path: ADMIN_ROLES_EDITHTML, H: http.HandlerFunc(controller.RolEditarRender)},
		{Path: ADMIN_ROLES_CREAR, H: http.HandlerFunc(controller.RolCrear)},
		{Path: ADMIN_ROLES_EDITAR, H: http.HandlerFunc(controller.RolEditar)},
	}
}
