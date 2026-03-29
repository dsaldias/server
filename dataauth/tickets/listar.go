package tickets

import (
	"context"
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/99designs/gqlgen/graphql"
)

func AllTickets(db *sql.DB, q model.QueryTickets) ([]*model.RespTickets, error) {
	sql := `
	SELECT 
		t.id,
		t.usuario_id,
		CONCAT(u.nombres, ' ', u.apellido1, ' ', IFNULL(u.apellido2, '')) AS cliente,
		t.problema,
		t.estado,
		t.fecha_registro,
		tr.respuesta,
		CONCAT(u2.nombres, ' ', u2.apellido1, ' ', IFNULL(u2.apellido2, '')) AS soporte,
		u2.id AS soporte_id,
		tr.fecha_registro AS respondido
	FROM rbac_tickets t
	INNER JOIN rbac_usuarios u ON u.id = t.usuario_id
	LEFT JOIN (
		-- Subconsulta para obtener la última respuesta por ticket
		SELECT 
			tr1.tickets_id,
			tr1.respuesta,
			tr1.usuario_id,
			tr1.fecha_registro
		FROM rbac_tickets_respuestas tr1
		INNER JOIN (
			-- Identifica el último registro por tickets_id
			SELECT 
				tickets_id,
				MAX(fecha_registro) AS ultima_fecha
			FROM rbac_tickets_respuestas
			GROUP BY tickets_id
		) tr2 ON tr1.tickets_id = tr2.tickets_id 
			AND tr1.fecha_registro = tr2.ultima_fecha
	) tr ON t.id = tr.tickets_id
	LEFT JOIN rbac_usuarios u2 ON u2.id = tr.usuario_id 
	order by t.id desc
	`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fs := []*model.RespTickets{}
	for rows.Next() {
		r := model.RespTickets{}
		er := parseRow(rows, &r)
		if er != nil {
			return nil, er
		}
		fs = append(fs, &r)
	}

	return fs, nil
}

func MisTickets(db *sql.DB, userid string) ([]*model.RespTickets, error) {
	sql := `
	SELECT 
		t.id,
		t.usuario_id,
		CONCAT(u.nombres, ' ', u.apellido1, ' ', IFNULL(u.apellido2, '')) AS cliente,
		t.problema,
		t.estado,
		t.fecha_registro,
		tr.respuesta,
		CONCAT(u2.nombres, ' ', u2.apellido1, ' ', IFNULL(u2.apellido2, '')) AS soporte,
		u2.id AS soporte_id,
		tr.fecha_registro AS respondido
	FROM rbac_tickets t
	INNER JOIN rbac_usuarios u ON u.id = t.usuario_id
	LEFT JOIN (
		-- Subconsulta para obtener la última respuesta por ticket
		SELECT 
			tr1.tickets_id,
			tr1.respuesta,
			tr1.usuario_id,
			tr1.fecha_registro
		FROM rbac_tickets_respuestas tr1
		INNER JOIN (
			-- Identifica el último registro por tickets_id
			SELECT 
				tickets_id,
				MAX(fecha_registro) AS ultima_fecha
			FROM rbac_tickets_respuestas
			GROUP BY tickets_id
		) tr2 ON tr1.tickets_id = tr2.tickets_id 
			AND tr1.fecha_registro = tr2.ultima_fecha
	) tr ON t.id = tr.tickets_id
	LEFT JOIN rbac_usuarios u2 ON u2.id = tr.usuario_id
	where t.usuario_id = ?
	order by t.id desc
	`
	rows, err := db.Query(sql, userid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fs := []*model.RespTickets{}
	for rows.Next() {
		r := model.RespTickets{}
		er := parseRow(rows, &r)
		if er != nil {
			return nil, er
		}
		fs = append(fs, &r)
	}

	return fs, nil
}

func Get(ctx context.Context, db *sql.DB, id string) (*model.Ticket, error) {
	fields := graphql.CollectFieldsCtx(ctx, nil)

	selectedFields := map[string]bool{}
	for _, field := range fields {
		selectedFields[field.Name] = true
	}

	sql := `
	select 
	t.id,
	t.usuario_id,
	t.problema,
	t.estado,
	t.fecha_registro
	from rbac_tickets t 
	where t.id = ?
	`
	row := db.QueryRow(sql, id)

	r := model.Ticket{}
	er := parseRow2(row, &r)
	if er != nil {
		return nil, er
	}
	r.Respuestas = []*model.TicketsRespuestas{}
	if selectedFields["respuestas"] {
		r.Respuestas, er = Respuestas(db, id)
		if er != nil {
			return nil, er
		}
	}

	return &r, nil
}
