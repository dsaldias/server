package tickets

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"
)

func Respuestas(db *sql.DB, idticket string) ([]*model.TicketsRespuestas, error) {
	sql := `select id, tickets_id,usuario_id,respuesta, fecha_registro from rbac_tickets_respuestas where tickets_id=?`
	rows, err := db.Query(sql, idticket)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ts := []*model.TicketsRespuestas{}
	for rows.Next() {
		t := model.TicketsRespuestas{}
		er := parseRowR(rows, &t)
		if er != nil {
			return nil, er
		}
		ts = append(ts, &t)
	}

	return ts, nil
}
