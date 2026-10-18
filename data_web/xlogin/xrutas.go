package xlogin

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/data_web/avisos"
	"github.com/dsaldias/server/data_web/roles"
	"github.com/dsaldias/server/data_web/tickets"
	"github.com/dsaldias/server/data_web/unidades"
	"github.com/dsaldias/server/data_web/usuarios"
	"github.com/dsaldias/server/data_web/xchat"
	"github.com/dsaldias/server/dataauth/utilidad"
)

var (
	ADMIN_HOME           = "/adminx/"
	ADMIN_LOGIN          = "/adminx/login"
	ADMIN_LOGOUT         = "/adminx/logout"
	ADMIN_INICIO         = "/adminx/inicio"
	ADMIN_CONTEXT        = "/adminx/context"
	ADMIN_PERFIL         = "/adminx/perfil"
	ADMIN_CONECTADOS     = "/adminx/conectados"
	ADMIN_AVISOS_ACTIVOS = "/adminx/avisos/activos"
)

// RutasAdmin registra el panel base y cualquier módulo adicional del
// consumidor. Los módulos adicionales pasan por el mismo envoltorio que los
// módulos propios, por lo que una navegación parcial y una actualización
// completa conservan el layout administrativo.
func RutasAdmin(db *sql.DB, conf Config, extraRoutes ...*utilidad.Handlers2) []*utilidad.Handlers2 {
	controller := Controller{DB: db, Configur: &conf}
	rutas := []*utilidad.Handlers2{}
	rutas_usuarios := usuarios.Rutas(db)
	rutas_roles := roles.Rutas(db)
	rutas_unidades := unidades.Rutas(db)
	rutas_avisos := avisos.Rutas(db)
	rutas_tickets := tickets.Rutas(db)
	rutas_chat := xchat.Rutas(db)

	rutas = append(rutas, &utilidad.Handlers2{Path: "/adminx", H: http.HandlerFunc(controller.LoginRender)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_HOME, H: http.HandlerFunc(controller.LoginRender)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_LOGIN, H: http.HandlerFunc(controller.LoginSubmit)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_LOGOUT, H: http.HandlerFunc(controller.Logout)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_INICIO, H: http.HandlerFunc(controller.InicioRender)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_CONTEXT, H: http.HandlerFunc(controller.ContextSubmit)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_PERFIL, H: http.HandlerFunc(controller.Perfil)})
	rutas = append(rutas, rutas_chat...)
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_CONECTADOS, H: http.HandlerFunc(controller.ConnectedUsers)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_AVISOS_ACTIVOS, H: http.HandlerFunc(controller.ActiveAlerts)})

	moduleRoutes := append(rutas_usuarios, rutas_roles...)
	moduleRoutes = append(moduleRoutes, rutas_unidades...)
	moduleRoutes = append(moduleRoutes, rutas_avisos...)
	moduleRoutes = append(moduleRoutes, rutas_tickets...)

	// estas rutas deben ir al final para dar prioridad a las rutas del consumidor
	moduleRoutes = append(moduleRoutes, extraRoutes...)
	for _, moduleRoute := range moduleRoutes {
		route := moduleRoute
		rutas = append(rutas, &utilidad.Handlers2{
			Path: route.Path,
			H:    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { controller.renderModuleRoute(w, r, route.H) }),
		})
	}

	return rutas
}
