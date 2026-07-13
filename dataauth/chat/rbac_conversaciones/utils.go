package rbacconversaciones

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func parseRow(r *sql.Row, t *model.ChatConversacion) error {
	return r.Scan(
		&t.ID,
		&t.Tipo,
		&t.Nombre,
		&t.FotoURL,
		&t.CreatedBy,
		&t.FechaRegistro,
		&t.FechaUpdate,
	)
}

func parseRows(r *sql.Rows, t *model.ResponseChatConversacion) error {
	return r.Scan(
		&t.ID,
		&t.Tipo,
		&t.Nombre,
		&t.FotoURL,
		&t.CreatedBy,
		&t.FechaRegistro,
		&t.FechaUpdate,
		&t.NoLeidos,
		&t.UltimoMensaje,
	)
}
