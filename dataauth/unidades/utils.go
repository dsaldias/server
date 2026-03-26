package unidades

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func parse(r *sql.Row, t *model.Unidad) error {
	return r.Scan(
		&t.ID,
		&t.Nombre,
		&t.Descripcion,
		&t.Orden,
		&t.Latitud,
		&t.Longitud,
		&t.FechaRegistro,
	)
}

func parseRows(r *sql.Rows, t *model.Unidad) error {
	return r.Scan(
		&t.ID,
		&t.Nombre,
		&t.Descripcion,
		&t.Orden,
		&t.Latitud,
		&t.Longitud,
		&t.FechaRegistro,
	)
}
