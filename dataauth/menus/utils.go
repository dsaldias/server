package menus

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func parseRows(rows *sql.Rows, t *model.Menus) error {
	return rows.Scan(
		&t.ID,
		&t.Label,
		&t.Path,
		&t.Icon,
		&t.Color,
		&t.Grupo,
		&t.Orden,
	)
}
