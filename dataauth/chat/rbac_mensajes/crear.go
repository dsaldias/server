package rbacmensajes

import (
	"database/sql"
	"strconv"

	"github.com/dsaldias/server/graph_auth/model"
)

func Crear(db *sql.DB, input model.NewChatMensaje) (*model.ChatMensaje, error) {
	sq := `insert into rbac_mensajes(conversacion_id,sender_id,tipo,texto) values(?,?,?,?)`
	res, err := db.Exec(sq, input.ConversacionID, input.SenderID, input.Tipo, input.Texto)
	if err != nil {
		return nil, err
	}

	idd, _ := res.LastInsertId()
	id := strconv.FormatInt(idd, 10)
	return Get(db, id)
}
