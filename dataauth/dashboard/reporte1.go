package dashboard

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func Reporte1(db *sql.DB) ([]*model.ResponseReporte1, error) {
	sql := `
	select 
	'usuarios' as nombre,
	count(u.id) as valor
	from rbac_usuarios u 
	union all
	select
	'roles' as nombre,
	count(r.id) as valor
	from rbac_roles r
	union all 
	select 
	'menus' as nombre,
	count(m.id) as valor
	from rbac_menus m 
	union all
	select 
	'unidades' as nombre,
	count(un.id) as valor
	from rbac_unidades un
	union all
	SELECT 
	'usuarios diarios' as nombre,
	COUNT(u.id) AS valor
	FROM rbac_usuarios u 
	WHERE u.last_login >= CURDATE()
	union all
	SELECT 
	'usuarios mensual' as nombre,
	COUNT(u.id) AS valor
	FROM rbac_usuarios u 
	WHERE u.last_login >= NOW() - INTERVAL 1 MONTH
	union all
	SELECT 
	'tickets pendientes' as nombre,
	COUNT(t.id) AS valor
	FROM rbac_tickets t 
	WHERE t.estado != "cerrado"
	`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []*model.ResponseReporte1{}
	for rows.Next() {
		r := model.ResponseReporte1{}
		er := parse1(rows, &r)
		if er != nil {
			return nil, er
		}
		res = append(res, &r)
	}

	return res, nil
}
