package rbacconversaciones

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

/* func Crear(db *sql.DB, input model.NewChatConversacion) (*model.ChatConversacion, error) {
	sq := `insert into rbac_conversaciones(tipo,nombre,created_by) values(?,?,?)`
	res, err := db.Exec(sq, input.Tipo, input.Nombre, input.CreatedBy)
	if err != nil {
		return nil, err
	}
	idd, _ := res.LastInsertId()
	id := strconv.FormatInt(idd, 10)
	return Get(db, id)
} */

func CrearConMiembros(tx *sql.Tx, input model.NewChatConversacion) (int64, error) {
	sq := `insert into rbac_conversaciones(tipo,nombre,created_by) values(?,?,?)`
	res, err := tx.Exec(sq, input.Tipo, input.Nombre, input.CreatedBy)
	if err != nil {
		return -1, err
	}
	idd, _ := res.LastInsertId()

	sq = `insert into rbac_conversaciones_miembros(conversacion_id,usuario_id) value(?,?)`
	res, err = tx.Exec(sq, idd, input.CreatedBy)
	if err != nil {
		return -1, err
	}

	sq = `insert into rbac_conversaciones_miembros(conversacion_id,usuario_id) value(?,?)`
	res, err = tx.Exec(sq, idd, input.DestinatorID)
	if err != nil {
		return -1, err
	}

	return idd, nil
}
