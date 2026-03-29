package dashboard

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func Reporte2(db *sql.DB) ([]*model.ResponseReporte2, error) {
	sql := `
	SELECT DATE(u.last_login) AS fecha, COUNT(u.id) AS rbac_usuarios
	FROM rbac_usuarios u 
	WHERE u.last_login >= NOW() - INTERVAL 1 MONTH
	GROUP BY DATE(u.last_login)
	ORDER BY fecha
	`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []*model.ResponseReporte2{}
	for rows.Next() {
		r := model.ResponseReporte2{}
		er := parse2(rows, &r)
		if er != nil {
			return nil, er
		}
		formato := "02/01/2006"
		r.FechaStr = r.Fecha.Format(formato)

		res = append(res, &r)
	}

	return res, nil
}

func Reporte2b(db *sql.DB) ([]*model.ResponseReporte2b, error) {
	sql := `
	SELECT 
	DATE_FORMAT(expire, '%Y-%m') AS mes, 
	COUNT(DISTINCT usuario_id) AS rbac_usuarios
	FROM rbac_session_keys
	WHERE expire >= NOW() - INTERVAL 1 YEAR
	GROUP BY DATE_FORMAT(expire, '%Y-%m')
	ORDER BY mes
	`

	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []*model.ResponseReporte2b{}
	for rows.Next() {
		r := model.ResponseReporte2b{}
		er := parse2b(rows, &r)
		if er != nil {
			return nil, er
		}

		res = append(res, &r)
	}

	return res, nil
}
