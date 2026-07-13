package rbacmensajes

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func parseRow(r *sql.Row, t *model.ChatMensaje) error {
	return r.Scan(
		&t.ID,
		&t.ConversacionID,
		&t.SenderID,
		&t.Tipo,
		&t.Texto,
		&t.CreatedAt,
		&t.EditedAt,
		&t.DeletedAt,
	)
}

func parseRows(r *sql.Rows, t *model.ChatMensaje) error {
	return r.Scan(
		&t.ID,
		&t.ConversacionID,
		&t.SenderID,
		&t.Tipo,
		&t.Texto,
		&t.CreatedAt,
		&t.EditedAt,
		&t.DeletedAt,
	)
}
