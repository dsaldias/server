package notificaciones

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func Get(db *sql.DB, id string) (*model.Notificacion, error) {
	sql := `select id, mensaje,creado_por_id,desde,hasta,fecha_registro from notificaciones where id=?`
	row := db.QueryRow(sql, id)
	r := model.Notificacion{}
	err := parseRow(row, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func GetNotificacionesActivas(db *sql.DB) ([]*model.Notificacion, error) {
	sql := `select id, mensaje,creado_por_id,desde,hasta,fecha_registro from notificaciones where now() between desde and hasta`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rs := []*model.Notificacion{}
	for rows.Next() {
		r := model.Notificacion{}
		er := parseRows(rows, &r)
		if er != nil {
			return nil, err
		}

		rs = append(rs, &r)
	}

	return rs, nil
}
