package unidades

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func GetById(db *sql.DB, id string) (*model.Unidad, error) {
	sql := `select id, nombre, descripcion, orden,ST_X(ubicacion) AS latitud, ST_Y(ubicacion) AS longitud, fecha_registro from rbac_unidades where id=?`
	row := db.QueryRow(sql, id)
	u := model.Unidad{}

	err := parse(row, &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetFirtsByUser(db *sql.DB, userid string) (*model.Unidad, error) {
	sql := `
	select un.id, un.nombre, un.descripcion, un.orden,ST_X(un.ubicacion) AS latitud, ST_Y(un.ubicacion) AS longitud, un.fecha_registro 
	from rbac_unidades un
	inner join rbac_rol_usuario_unidades ruu on ruu.unidad_id = un.id 
	where ruu.usuario_id = ?
	order by un.id asc
	limit 1
	`
	row := db.QueryRow(sql, userid)
	u := model.Unidad{}

	err := parse(row, &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func Listar(db *sql.DB) ([]*model.Unidad, error) {
	sql := `select id, nombre, descripcion, orden,ST_X(ubicacion) AS latitud, ST_Y(ubicacion) AS longitud, fecha_registro from rbac_unidades order by id, orden`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rs := []*model.Unidad{}
	for rows.Next() {
		u := model.Unidad{}
		er := parseRows(rows, &u)
		if er != nil {
			return nil, er
		}
		rs = append(rs, &u)
	}

	return rs, nil
}
