package rbacstatus

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func parseRow(r *sql.Row, t *model.ChatMensajeStatus) error {
	return r.Scan(
		&t.ID,
		&t.MessageID,
		&t.UsuarioID,
		&t.Status,
		&t.UpdatedAt,
	)
}
