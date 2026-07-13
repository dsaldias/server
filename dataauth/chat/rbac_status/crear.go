package rbacstatus

import (
	"database/sql"
	"strconv"

	"github.com/dsaldias/server/graph_auth/model"
)

func Crear(db *sql.DB, input model.NewChatMensajeStatus) (*model.ChatMensajeStatus, error) {
	sq := `insert into rbac_mensaje_status(message_id,usuario_id,status) values(?,?,?)`
	res, err := db.Exec(sq, input.MessageID, input.UsuarioID, input.Status)
	if err != nil {
		return nil, err
	}
	idd, _ := res.LastInsertId()
	id := strconv.FormatInt(idd, 10)
	return Get(db, id)
}
