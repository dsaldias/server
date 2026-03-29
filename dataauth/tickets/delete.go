package tickets

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/xnotificaciones"
)

func Cerrar(ctx context.Context, db *sql.DB, id string) (*model.Ticket, error) {
	sql := `update rbac_tickets set estado='cerrado' where id = ?`
	_, err := db.Exec(sql, id)
	if err != nil {
		return nil, err
	}

	m := fmt.Sprintf("el ticket: %s fue cerrado.", id)
	data := datanotify("red")
	xnotificaciones.EnviarNotificacion(context.Background(), m, &data)
	return Get(ctx, db, id)
}
