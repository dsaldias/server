package rbacconversaciones

import (
	"database/sql"
	"errors"

	rbacconversacionesmiembros "github.com/dsaldias/server/dataauth/chat/rbac_conversaciones_miembros"
	"github.com/dsaldias/server/graph_auth/model"
)

func Get(db *sql.DB, id string) (*model.ChatConversacion, error) {
	sq := `select id,tipo,nombre,foto_url,created_by,fecha_registro,fecha_update from rbac_conversaciones where id=?`
	row := db.QueryRow(sq, id)
	r := model.ChatConversacion{}
	err := parseRow(row, &r)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("chat conversacion no encontrado")
		}
		return nil, err
	}

	return &r, nil
}

func ByUser(db *sql.DB, miembroID string) ([]*model.ResponseChatConversacion, error) {
	sq := `
		SELECT
			c.id,
			c.tipo,
			COALESCE(
				(
					SELECT concat(u.nombres,' ',u.apellido1)
					FROM rbac_conversaciones_miembros cm_otro
					INNER JOIN rbac_usuarios u
						ON u.id = cm_otro.usuario_id
					WHERE cm_otro.conversacion_id = c.id
						AND cm_otro.usuario_id <> cm.usuario_id
						AND cm_otro.left_at IS NULL
					LIMIT 1
				),
				c.nombre
			) AS nombre,
			c.foto_url,
			c.created_by,
			c.fecha_registro,
			c.fecha_update,
			COALESCE(
				SUM(
					IF(
						m.id IS NOT NULL
						AND m.sender_id <> cm.usuario_id
						AND m.deleted_at IS NULL
						AND (
							ms.id IS NULL
							OR ms.status <> 'read'
						),
						1,
						0
					)
				),
				0
			) AS no_leidos,
			(
        SELECT m.texto
        FROM rbac_mensajes m
        WHERE m.conversacion_id = c.id
          AND m.deleted_at IS NULL
        ORDER BY m.id DESC
        LIMIT 1
    	) AS ultimo_mensaje
		FROM rbac_conversaciones c

		INNER JOIN rbac_conversaciones_miembros cm
			ON cm.conversacion_id = c.id
			AND cm.usuario_id = ?
			AND cm.left_at IS NULL

		LEFT JOIN rbac_mensajes m
			ON m.conversacion_id = c.id
			AND m.deleted_at IS NULL

		LEFT JOIN rbac_mensaje_status ms
			ON ms.message_id = m.id
			AND ms.usuario_id = cm.usuario_id

		GROUP BY
			c.id,
			c.tipo,
			c.nombre,
			c.foto_url,
			c.created_by,
			c.fecha_registro,
			c.fecha_update,
			cm.usuario_id

		ORDER BY c.fecha_update DESC
	`

	rows, err := db.Query(sq, miembroID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rs := []*model.ResponseChatConversacion{}

	for rows.Next() {
		r := model.ResponseChatConversacion{}

		if err := parseRows(rows, &r); err != nil {
			return nil, err
		}

		r.Miembros, err = rbacconversacionesmiembros.ByConversacion(db, r.ID)
		if err != nil {
			return nil, err
		}

		rs = append(rs, &r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rs, nil
}

func BuscarConversacionPrivada(db *sql.DB, sender_id, destinator_id string) (int64, error) {
	sq := `
	SELECT c.id
	FROM rbac_conversaciones c
	JOIN rbac_conversaciones_miembros cm
			ON cm.conversacion_id = c.id
	WHERE c.tipo = 'privado'
		AND cm.left_at IS NULL
		AND cm.usuario_id IN (?, ?)
	GROUP BY c.id
	HAVING COUNT(DISTINCT cm.usuario_id) = 2
		AND COUNT(*) = 2
	LIMIT 1;
	`
	encontrado_id := int64(0)
	row := db.QueryRow(sq, sender_id, destinator_id)
	err := row.Scan(&encontrado_id)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, nil
		}
		return -1, err
	}

	return encontrado_id, nil
}
