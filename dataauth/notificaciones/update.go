package notificaciones

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func Actualizar(db *sql.DB, input model.UpdNotificacion, userid string) (*model.Notificacion, error) {
	sql := `update rbac_notificaciones set mensaje=?, creado_por_id=?, desde=?, hasta=? where id=?`
	_, err := db.Exec(sql, input.Mensaje, userid, input.Desde, input.Hasta, input.ID)
	if err != nil {
		return nil, err
	}

	return Get(db, input.ID)
}
