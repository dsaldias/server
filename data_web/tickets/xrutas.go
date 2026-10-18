package tickets

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

const (
	ADMIN_TICKETS           = "/adminx/tickets"
	ADMIN_TICKETS_LISTAHTML = ADMIN_TICKETS + "/listahtml"
	ADMIN_TICKETS_FORMHTML  = ADMIN_TICKETS + "/formulariohtml"
	ADMIN_TICKETS_VERHTML   = ADMIN_TICKETS + "/visualizarhtml"
	ADMIN_TICKETS_CREAR     = ADMIN_TICKETS + "/crear"
	ADMIN_TICKETS_RESPONDER = ADMIN_TICKETS + "/responder"
	ADMIN_TICKETS_CERRAR    = ADMIN_TICKETS + "/cerrar"
)

func Rutas(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{DB: db}
	return []*utilidad.Handlers2{
		{Path: ADMIN_TICKETS, H: http.HandlerFunc(controller.TicketsRender)},
		{Path: ADMIN_TICKETS_LISTAHTML, H: http.HandlerFunc(controller.TicketsListaHTML)},
		{Path: ADMIN_TICKETS_FORMHTML, H: http.HandlerFunc(controller.TicketFormularioRender)},
		{Path: ADMIN_TICKETS_VERHTML, H: http.HandlerFunc(controller.TicketVisualizarRender)},
		{Path: ADMIN_TICKETS_CREAR, H: http.HandlerFunc(controller.TicketCrear)},
		{Path: ADMIN_TICKETS_RESPONDER, H: http.HandlerFunc(controller.TicketResponder)},
		{Path: ADMIN_TICKETS_CERRAR, H: http.HandlerFunc(controller.TicketCerrar)},
	}
}
