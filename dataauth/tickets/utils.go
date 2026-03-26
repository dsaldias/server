package tickets

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/xnotificaciones"
)

func parseRow(r *sql.Rows, t *model.RespTickets) error {
	return r.Scan(
		&t.ID,
		&t.UsuarioID,
		&t.Usuario,
		&t.Problema,
		&t.Estado,
		&t.FechaRegistro,
		&t.Respuesta,
		&t.Soporte,
		&t.SoporteID,
		&t.Respondido,
	)
}

func parseRow2(r *sql.Row, t *model.Ticket) error {
	return r.Scan(
		&t.ID,
		&t.UsuarioID,
		&t.Problema,
		&t.Estado,
		&t.FechaRegistro,
	)
}

func parseRowR(r *sql.Rows, t *model.TicketsRespuestas) error {
	return r.Scan(
		&t.ID,
		&t.TicketsID,
		&t.UsuarioID,
		&t.Respuesta,
		&t.FechaRegistro,
	)
}

func datanotify(color string) xnotificaciones.DataNotify {
	d := xnotificaciones.DataNotify{Color: &color}
	return d
}
