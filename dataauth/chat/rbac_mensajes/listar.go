package rbacmensajes

import (
	"database/sql"
	"errors"

	"github.com/dsaldias/server/graph_auth/model"
)

func Get(db *sql.DB, id string) (*model.ChatMensaje, error) {
	sq := `select id,conversacion_id,sender_id,tipo,texto,created_at,edited_at,deleted_at from rbac_mensajes where id=?`
	row := db.QueryRow(sq, id)
	r := model.ChatMensaje{}
	err := parseRow(row, &r)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("chat mensajes no encontrado")
		}
		return nil, err
	}
	return &r, nil
}

func ByChat(db *sql.DB, conversacionID string, miembros []*model.ChatConversacionMiembro) ([]*model.ChatMensaje, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Registrar o actualizar como read todos los mensajes
	// para cada miembro de la conversación.
	const statusSQL = `
		INSERT INTO rbac_mensaje_status (
			message_id,
			usuario_id,
			status
		)
		SELECT
			m.id,
			?,
			'read'
		FROM rbac_mensajes m
		WHERE m.conversacion_id = ?
		  AND m.deleted_at IS NULL
		ON DUPLICATE KEY UPDATE
			status = 'read',
			updated_at = CURRENT_TIMESTAMP
	`

	for _, miembro := range miembros {
		if miembro == nil || miembro.UsuarioID == "" {
			continue
		}

		_, err := tx.Exec(
			statusSQL,
			miembro.UsuarioID,
			conversacionID,
		)
		if err != nil {
			return nil, err
		}
	}

	const mensajesSQL = `
		SELECT
			id,
			conversacion_id,
			sender_id,
			tipo,
			texto,
			created_at,
			edited_at,
			deleted_at
		FROM rbac_mensajes
		WHERE conversacion_id = ?
		ORDER BY id DESC
		LIMIT 50
	`

	rows, err := tx.Query(mensajesSQL, conversacionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rs := make([]*model.ChatMensaje, 0, 50)

	for rows.Next() {
		r := model.ChatMensaje{}

		if err := parseRows(rows, &r); err != nil {
			return nil, err
		}

		rs = append(rs, &r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return rs, nil
}
