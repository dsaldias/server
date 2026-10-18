package avisos

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

const (
	ADMIN_NOTIFICACIONES   = "/adminx/avisos"
	ADMIN_AVISOS_LISTAHTML = ADMIN_NOTIFICACIONES + "/listahtml"
	ADMIN_AVISOS_FORMHTML  = ADMIN_NOTIFICACIONES + "/formulariohtml"
	ADMIN_AVISOS_VERHTML   = ADMIN_NOTIFICACIONES + "/visualizarhtml"
	ADMIN_AVISOS_EDITHTML  = ADMIN_NOTIFICACIONES + "/editarhtml"
	ADMIN_AVISOS_CREAR     = ADMIN_NOTIFICACIONES + "/crear"
	ADMIN_AVISOS_EDITAR    = ADMIN_NOTIFICACIONES + "/editar"
)

func Rutas(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{DB: db}
	return []*utilidad.Handlers2{
		{Path: ADMIN_NOTIFICACIONES, H: http.HandlerFunc(controller.AvisosRender)},
		{Path: ADMIN_AVISOS_LISTAHTML, H: http.HandlerFunc(controller.AvisosListaHTML)},
		{Path: ADMIN_AVISOS_FORMHTML, H: http.HandlerFunc(controller.AvisoFormularioRender)},
		{Path: ADMIN_AVISOS_VERHTML, H: http.HandlerFunc(controller.AvisoVisualizarRender)},
		{Path: ADMIN_AVISOS_EDITHTML, H: http.HandlerFunc(controller.AvisoEditarRender)},
		{Path: ADMIN_AVISOS_CREAR, H: http.HandlerFunc(controller.AvisoCrear)},
		{Path: ADMIN_AVISOS_EDITAR, H: http.HandlerFunc(controller.AvisoEditar)},
	}
}
