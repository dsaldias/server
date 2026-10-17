package tickets

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

const ADMIN_TICKETS = "/adminx/tickets"

func Rutas(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{DB: db}
	return []*utilidad.Handlers2{{Path: ADMIN_TICKETS, H: http.HandlerFunc(controller.TicketsRender)}}
}
