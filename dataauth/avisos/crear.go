package avisos

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/dsaldias/server/dataauth/xnotificaciones"
	"github.com/dsaldias/server/graph_auth/model"
)

func Crear(ctx context.Context, db *sql.DB, input model.NewNotificacion, userid string) (*model.Notificacion, error) {
	sql := `insert into rbac_notificaciones(mensaje,creado_por_id,desde,hasta) values(?,?,?,?)`
	res, err := db.Exec(sql, input.Mensaje, userid, input.Desde, input.Hasta)
	if err != nil {
		return nil, err
	}
	idd, _ := res.LastInsertId()
	id := strconv.FormatInt(idd, 10)

	tipo := "alerta"
	mapa := map[string]any{}
	// mapa["mensaje"] = input.Mensaje

	d := xnotificaciones.DataNotify{Tipo: &tipo, Datos: mapa}
	xnotificaciones.EnviarNotificacion(ctx, input.Mensaje, &d)

	return Get(db, id)
}
