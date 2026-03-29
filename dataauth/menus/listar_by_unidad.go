package menus

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func ListarByUserUnidad(db *sql.DB, input model.InputMe, userid string) ([]*model.Menus, error) {
	xsql := `
	SELECT distinct m.id, m.label, m.path, m.icon, m.color, m.grupo, m.orden, m.padre_id
	FROM menus m
	LEFT JOIN menus_usuario mu ON mu.menu_id = m.id
	LEFT JOIN rol_menus rm ON rm.menu_id = m.id
	LEFT JOIN roles r ON r.id = rm.rol_id 
	LEFT JOIN rol_usuario_unidades ruu ON ruu.rol_id = r.id
	WHERE (ruu.usuario_id = ? AND ruu.unidad_id = ?)
	OR (mu.usuario_id = ?)
	ORDER BY m.grupo, m.orden, m.id;
	`

	rows, err := db.Query(xsql, userid, input.UnidadID, userid)
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
