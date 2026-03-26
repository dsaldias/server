package permisos

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func parse(rows *sql.Rows, t *model.ResponsePermisoMe) error {
	return rows.Scan(
		&t.Metodo,
		&t.Nombre,
		&t.Descripcion,
		&t.Grupo,
		&t.FechaRegistro,
		&t.FechaAsignado,
	)
}

func parseRows(rows *sql.Rows, t *model.Permiso) error {
	return rows.Scan(
		&t.Metodo,
		&t.Nombre,
		&t.Descripcion,
		&t.Grupo,
		&t.FechaRegistro,
	)
}
