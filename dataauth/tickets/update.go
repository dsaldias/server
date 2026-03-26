package tickets

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/xnotificaciones"
)

func Update(ctx context.Context, db *sql.DB, input model.NewTicketRespuesta, userid string) (*model.Ticket, error) {
	t, err := Get(ctx, db, input.TicketsID)
	if err != nil {
		return nil, err
	}
	if t.Estado == "cerrado" {
		return nil, errors.New("el ticket esta cerrado")
	}
	estado := "cliente"
	if t.UsuarioID != userid {
		estado = "soporte"
	}

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	sql := `insert into tickets_respuestas(tickets_id,usuario_id,respuesta) values(?,?,?)`
	_, err = tx.Exec(sql, input.TicketsID, userid, input.Respuesta)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	sql = `update tickets set estado=? where id=?`
	_, err = tx.Exec(sql, estado, input.TicketsID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	m := fmt.Sprintf("ticket: %s respondido", input.TicketsID)
	data := datanotify("purple")
	xnotificaciones.EnviarNotificacion(context.Background(), m, &data)
	return Get(ctx, db, input.TicketsID)
}
