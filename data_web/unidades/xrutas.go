package unidades

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

const (
	ADMIN_UNIDADES           = "/adminx/unidades"
	ADMIN_UNIDADES_LISTAHTML = ADMIN_UNIDADES + "/listahtml"
	ADMIN_UNIDADES_FORMHTML  = ADMIN_UNIDADES + "/formulariohtml"
	ADMIN_UNIDADES_VERHTML   = ADMIN_UNIDADES + "/visualizarhtml"
	ADMIN_UNIDADES_EDITHTML  = ADMIN_UNIDADES + "/editarhtml"
	ADMIN_UNIDADES_CREAR     = ADMIN_UNIDADES + "/crear"
	ADMIN_UNIDADES_EDITAR    = ADMIN_UNIDADES + "/editar"
)

func Rutas(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{DB: db}
	return []*utilidad.Handlers2{
		{Path: ADMIN_UNIDADES, H: http.HandlerFunc(controller.UnidadesRender)},
		{Path: ADMIN_UNIDADES_LISTAHTML, H: http.HandlerFunc(controller.UnidadesListaHTML)},
		{Path: ADMIN_UNIDADES_FORMHTML, H: http.HandlerFunc(controller.UnidadFormularioRender)},
		{Path: ADMIN_UNIDADES_VERHTML, H: http.HandlerFunc(controller.UnidadVisualizarRender)},
		{Path: ADMIN_UNIDADES_EDITHTML, H: http.HandlerFunc(controller.UnidadEditarRender)},
		{Path: ADMIN_UNIDADES_CREAR, H: http.HandlerFunc(controller.UnidadCrear)},
		{Path: ADMIN_UNIDADES_EDITAR, H: http.HandlerFunc(controller.UnidadEditar)},
	}
}
