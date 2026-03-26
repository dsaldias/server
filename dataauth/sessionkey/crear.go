package sessionkey

import (
	"database/sql"
	"time"

	"github.com/dsaldias/server/graph_auth/model"
)

func CrearApikey(db *sql.DB, userid, apikey string, exp time.Time) (*model.SessionKey, error) {
	key, err := generarStringUnico()
	if err != nil {
		return nil, err
	}

	input := model.NewSessionKey{}
	input.UsuarioID = userid
	input.Apikey = apikey
	input.Key = key
	input.Expire = exp

	sql := "insert into session_keys(`usuario_id`,`key`,`apikey`,`expire`) values(?,?,?,?)"
	_, err = db.Exec(sql, input.UsuarioID, input.Key, input.Apikey, input.Expire)
	if err != nil {
		return nil, err
	}

	return GetyKey(db, input.Key)
}
