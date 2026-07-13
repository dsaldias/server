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
