package rbacstatus

import (
	"database/sql"
	"errors"

	"github.com/dsaldias/server/graph_auth/model"
)

func Get(db *sql.DB, id string) (*model.ChatMensajeStatus, error) {
	sq := `select id,message_id,usuario_id,status,updated_at from rbac_mensaje_status where id=?`
	row := db.QueryRow(sq, id)
	r := model.ChatMensajeStatus{}
	err := parseRow(row, &r)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no existe mensaje status")
		}
		return nil, err
	}
	return &r, nil
}

func NoLeidos(db *sql.DB, user_id string) (int32, error) {
	sq := `
	SELECT COUNT(*) AS mensajes_sin_leer
	FROM rbac_mensajes m
	INNER JOIN rbac_conversaciones_miembros cm
			ON cm.conversacion_id = m.conversacion_id
			AND cm.usuario_id = ?
			AND cm.left_at IS NULL
	LEFT JOIN rbac_mensaje_status ms
			ON ms.message_id = m.id
			AND ms.usuario_id = ?
	WHERE m.deleted_at IS NULL
		AND m.sender_id <> ?
		AND (
				ms.id IS NULL
				OR ms.status <> 'read'
		);
	`

	total := int32(0)
	row := db.QueryRow(sq, user_id, user_id, user_id)
	err := row.Scan(&total)
	if err != nil {
		return -1, err
	}

	return total, nil
}
