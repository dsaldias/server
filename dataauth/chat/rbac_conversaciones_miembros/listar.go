package rbacconversacionesmiembros

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func ByConversacion(db *sql.DB, conversacion_id string) ([]*model.ChatConversacionMiembro, error) {
	sq := `select id,conversacion_id,usuario_id,is_admin,joined_at,left_at,last_read_message_id from rbac_conversaciones_miembros where conversacion_id=?`
	rows, err := db.Query(sq, conversacion_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rs := []*model.ChatConversacionMiembro{}
	for rows.Next() {
		r := model.ChatConversacionMiembro{}
		er := parseRows(rows, &r)
		if er != nil {
			return nil, er
		}

		rs = append(rs, &r)
	}

	return rs, nil
}
