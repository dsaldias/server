package notificaciones

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func parseRow(r *sql.Row, t *model.Notificacion) error {
	return r.Scan(
		&t.ID,
		&t.Mensaje,
		&t.CreadoPorID,
		&t.Desde,
		&t.Hasta,
		&t.FechaRegistro,
	)
}

func parseRows(r *sql.Rows, t *model.Notificacion) error {
	return r.Scan(
		&t.ID,
		&t.Mensaje,
		&t.CreadoPorID,
		&t.Desde,
		&t.Hasta,
		&t.FechaRegistro,
	)
}
