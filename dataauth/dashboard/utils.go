package dashboard

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func parse1(r *sql.Rows, t *model.ResponseReporte1) error {
	return r.Scan(
		&t.Nombre,
		&t.Valor,
	)
}

func parse2(r *sql.Rows, t *model.ResponseReporte2) error {
	return r.Scan(
		&t.Fecha,
		&t.Valor,
	)
}

func parse2b(r *sql.Rows, t *model.ResponseReporte2b) error {
	return r.Scan(
		&t.Mes,
		&t.Valor,
	)
}
