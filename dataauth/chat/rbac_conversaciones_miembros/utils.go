package rbacconversacionesmiembros

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func parseRows(r *sql.Rows, t *model.ChatConversacionMiembro) error {
	return r.Scan(
		&t.ID,
		&t.ConversacionID,
		&t.UsuarioID,
		&t.IsAdmin,
		&t.JoinedAt,
		&t.LeftAt,
		&t.LastReadMessageID,
	)
}
