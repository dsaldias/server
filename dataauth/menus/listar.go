package menus

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func GetMenusbyRol(db *sql.DB, rol_id string) ([]*model.Menus, error) {
	sql := `
	select m.id,m.label,m.path,m.icon,m.color,m.grupo, m.orden 
	from menus m
	inner join rol_menus rm on rm.menu_id = m.id
	where rm.rol_id = ?
	`
	rows, err := db.Query(sql, rol_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mns := []*model.Menus{}
	for rows.Next() {
		m := model.Menus{}
		er := parseRows(rows, &m)
		if er != nil {
			return nil, er
		}
		mns = append(mns, &m)
	}
	return mns, nil
}

func Listar(db *sql.DB) ([]*model.Menus, error) {
	sql := `select id,label,path,icon,color,grupo, orden from menus order by grupo,orden,id`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mns := []*model.Menus{}
	for rows.Next() {
		m := model.Menus{}
		er := parseRows(rows, &m)
		if er != nil {
			return nil, er
		}
		mns = append(mns, &m)
	}
	return mns, nil
}

func MenusSueltos(db *sql.DB, userid string) ([]*model.Menus, error) {
	xsql := `
	SELECT m.id, m.label, m.path, m.icon, m.color, m.grupo, m.orden
	FROM menus m
	inner JOIN menus_usuario mu ON mu.menu_id = m.id
	WHERE mu.usuario_id = ?
	ORDER BY m.grupo, m.orden, m.id;
	`

	rows, err := db.Query(xsql, userid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mns := []*model.Menus{}
	for rows.Next() {
		m := model.Menus{}
		er := parseRows(rows, &m)
		if er != nil {
			return nil, er
		}
		mns = append(mns, &m)
	}
	return mns, nil
}
