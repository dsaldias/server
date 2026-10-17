package avisos

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

const ADMIN_NOTIFICACIONES = "/adminx/avisos"

func Rutas(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{}
	return []*utilidad.Handlers2{{Path: ADMIN_NOTIFICACIONES, H: http.HandlerFunc(controller.NotificacionesRender)}}
}
