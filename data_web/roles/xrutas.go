package roles

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

const ADMIN_ROLES = "/adminx/roles"

func Rutas(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{DB: db}
	return []*utilidad.Handlers2{{Path: ADMIN_ROLES, H: http.HandlerFunc(controller.RolesRender)}}
}
