package sessionkey

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func GetyKey(db *sql.DB, key string) (*model.SessionKey, error) {
	sql := `
	select sk.id, sk.usuario_id,sk.key,sk.apikey,sk.expire,sk.fecha_registro, u.estado 
	from rbac_session_keys sk 
	inner join rbac_usuarios u on u.id = sk.usuario_id 
	where sk.key= ?
	`
	row := db.QueryRow(sql, key)
	k := model.SessionKey{}

	err := parse(row, &k)
	if err != nil {
		return nil, err
	}
	return &k, nil
}
