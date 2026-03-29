package tickets

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/utils"
	"github.com/dsaldias/server/dataauth/xnotificaciones"
)

func Crear(ctx context.Context, db *sql.DB, input model.NewTicket, userid string) (*model.Ticket, error) {
	sql := `insert into rbac_tickets(usuario_id,problema) values(?,?)`
	res, err := db.Exec(sql, userid, input.Problema)
	if err != nil {
		return nil, err
	}
	idd, _ := res.LastInsertId()
	id := strconv.FormatInt(idd, 10)

	m := fmt.Sprintf("ticket: %s abierto", id)
	data := datanotify("purple")
	xnotificaciones.EnviarNotificacion(context.Background(), m, &data)

	utils.NotifyTicketCreated(db, id)

	return Get(ctx, db, id)
}
