package xlogin

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/data_web/avisos"
	"github.com/dsaldias/server/data_web/roles"
	"github.com/dsaldias/server/data_web/tickets"
	"github.com/dsaldias/server/data_web/unidades"
	"github.com/dsaldias/server/data_web/usuarios"
	"github.com/dsaldias/server/dataauth/utilidad"
)

var (
	ADMIN_HOME           = "/adminx/"
	ADMIN_LOGIN          = "/adminx/login"
	ADMIN_LOGOUT         = "/adminx/logout"
	ADMIN_INICIO         = "/adminx/inicio"
	ADMIN_CONTEXT        = "/adminx/context"
	ADMIN_PERFIL         = "/adminx/perfil"
	ADMIN_CHAT           = "/adminx/chat"
	ADMIN_CHAT_USERS     = "/adminx/chat/users"
	ADMIN_CHAT_COMPOSE   = "/adminx/chat/compose"
	ADMIN_CHAT_SEND      = "/adminx/chat/send"
	ADMIN_CONECTADOS     = "/adminx/conectados"
	ADMIN_AVISOS_ACTIVOS = "/adminx/avisos/activos"
)

func RutasAdmin(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{DB: db}
	rutas := []*utilidad.Handlers2{}
	rutas_usuarios := usuarios.Rutas(db)
	rutas_roles := roles.Rutas(db)
	rutas_unidades := unidades.Rutas(db)
	rutas_avisos := avisos.Rutas(db)
	rutas_tickets := tickets.Rutas(db)

	rutas = append(rutas, &utilidad.Handlers2{Path: "/adminx", H: http.HandlerFunc(controller.LoginRender)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_HOME, H: http.HandlerFunc(controller.LoginRender)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_LOGIN, H: http.HandlerFunc(controller.LoginSubmit)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_LOGOUT, H: http.HandlerFunc(controller.Logout)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_INICIO, H: http.HandlerFunc(controller.InicioRender)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_CONTEXT, H: http.HandlerFunc(controller.ContextSubmit)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_PERFIL, H: http.HandlerFunc(controller.Perfil)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_CHAT, H: http.HandlerFunc(controller.Chat)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_CHAT_USERS, H: http.HandlerFunc(controller.ChatUsers)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_CHAT_COMPOSE, H: http.HandlerFunc(controller.ChatCompose)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_CHAT_SEND, H: http.HandlerFunc(controller.ChatSend)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_CONECTADOS, H: http.HandlerFunc(controller.ConnectedUsers)})
	rutas = append(rutas, &utilidad.Handlers2{Path: ADMIN_AVISOS_ACTIVOS, H: http.HandlerFunc(controller.ActiveAlerts)})

	moduleRoutes := append(rutas_usuarios, rutas_roles...)
	moduleRoutes = append(moduleRoutes, rutas_unidades...)
	moduleRoutes = append(moduleRoutes, rutas_avisos...)
	moduleRoutes = append(moduleRoutes, rutas_tickets...)
	for _, moduleRoute := range moduleRoutes {
		route := moduleRoute
		rutas = append(rutas, &utilidad.Handlers2{
			Path: route.Path,
			H:    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { controller.renderModuleRoute(w, r, route.H) }),
		})
	}

	return rutas
}
