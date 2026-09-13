package dataadmin

import (
	"database/sql"
	"embed"
	"io/fs"
	"net/http"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/dataadmin/pages/avisos"
	"github.com/dsaldias/server/dataadmin/pages/mainlayout"
	xlogin "github.com/dsaldias/server/dataadmin/pages/mainlayout/login"
	"github.com/dsaldias/server/dataadmin/pages/roles"
	"github.com/dsaldias/server/dataadmin/pages/ticket"
	"github.com/dsaldias/server/dataadmin/pages/unidades"
	"github.com/dsaldias/server/dataadmin/pages/usuarios"
	"github.com/dsaldias/server/dataauth/utils"
)

//go:embed assets/*
var Assets embed.FS

var (
	WEB_PATH_BASE   = "/adminx/"
	WEB_PATH_LOGIN  = "/adminx/login"
	WEB_PATH_LOGOUT = "/adminx/logout"
	WEB_PATH_MAIN   = "/adminx/main"

	WEB_ADMIN_PATH_SET_COOKIE_UNIDAD = "/adminx/setcookie/data"
	WEB_ADMIN_PATH_USUARIOS          = "/adminx/usuarios"
	WEB_ADMIN_PATH_USUARIOS_FORMNEW  = "/adminx/usuarios/formnew"
	WEB_ADMIN_PATH_USUARIOS_CREAR    = "/adminx/usuarios/crear"
	WEB_ADMIN_PATH_USUARIOS_GET      = "/adminx/usuarios/{id}/get"
	WEB_ADMIN_PATH_USUARIOS_VER      = "/adminx/usuarios/{id}/ver"
	WEB_ADMIN_PATH_USUARIOS_EDIT     = "/adminx/usuarios/getperfil"
	WEB_ADMIN_PATH_USUARIOS_EDITSET  = "/adminx/usuarios/setperfil"

	WEB_ADMIN_PATH_ROLES         = "/adminx/roles"
	WEB_ADMIN_PATH_ROLES_FORMNEW = "/adminx/roles/formnew"
	WEB_ADMIN_PATH_ROLES_CREAR   = "/adminx/roles/crear"
	WEB_ADMIN_PATH_ROLES_GET     = "/adminx/roles/{id}/get"
	WEB_ADMIN_PATH_ROLES_VER     = "/adminx/roles/{id}/ver"

	WEB_ADMIN_PATH_UNIDADES         = "/adminx/unidades"
	WEB_ADMIN_PATH_UNIDADES_FORMNEW = "/adminx/unidades/formnew"
	WEB_ADMIN_PATH_UNIDADES_CREAR   = "/adminx/unidades/crear"
	WEB_ADMIN_PATH_UNIDADES_GET     = "/adminx/unidades/{id}/get"
	WEB_ADMIN_PATH_UNIDADES_VER     = "/adminx/unidades/{id}/ver"

	WEB_ADMIN_PATH_AVISOS         = "/adminx/avisos"
	WEB_ADMIN_PATH_AVISOS_FORMNEW = "/adminx/avisos/formnew"
	WEB_ADMIN_PATH_AVISOS_CREAR   = "/adminx/avisos/crear"
	WEB_ADMIN_PATH_AVISOS_GET     = "/adminx/avisos/{id}/get"
	WEB_ADMIN_PATH_AVISOS_VER     = "/adminx/avisos/{id}/ver"

	WEB_ADMIN_PATH_TICKETS               = "/adminx/tickets"
	WEB_ADMIN_PATH_TICKETS_FORMNEW       = "/adminx/tickets/formnew"
	WEB_ADMIN_PATH_TICKETS_CREAR         = "/adminx/tickets/crear"
	WEB_ADMIN_PATH_TICKETS_VER           = "/adminx/tickets/{id}/ver"
	WEB_ADMIN_PATH_TICKETS_RESPONDER_GET = "/adminx/tickets/{id}/get"
	WEB_ADMIN_PATH_TICKETS_RESPONDER     = "/adminx/tickets/responder"
)

func RutasFront(db *sql.DB) []*utils.Handlers2 {
	ssr := []*utils.Handlers2{}
	umain := mainlayout.MainController{DB: db}
	cont_login := xlogin.Logincontroller{DB: db}
	cont_main := mainlayout.MainController{DB: db}

	ucontroller := usuarios.UsuariosController{DB: db, C: &umain}
	rcontroller := roles.RolesController{DB: db, C: &umain}
	uncontroller := unidades.UnidadesController{DB: db, C: &umain}
	acontroller := avisos.AvisosController{DB: db, C: &umain}
	tcontroller := ticket.TicketController{DB: db, C: &umain}

	fs1, err := fs.Sub(Assets, "assets")
	if err != nil {
		panic(err)
	}
	templuiFS, err := fs.Sub(Assets, "assets/templui")
	if err != nil {
		panic(err)
	}

	ssr = append(ssr, &utils.Handlers2{Path: "/assets/*", H: http.StripPrefix("/assets/", http.FileServer(http.FS(fs1)))})
	ssr = append(ssr, &utils.Handlers2{Path: "/templui/*", H: http.StripPrefix("/templui/", http.FileServer(http.FS(templuiFS)))})

	ssr = append(ssr, &utils.Handlers2{Path: WEB_PATH_LOGOUT, H: http.HandlerFunc(cont_login.Logout)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_PATH_LOGIN, H: cont_login.Login()})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_PATH_BASE, H: templ.Handler(xlogin.Inicio())})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_PATH_MAIN, H: http.HandlerFunc(cont_main.MainLayout)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_SET_COOKIE_UNIDAD, H: http.HandlerFunc(umain.SetCookieUnidadId)})

	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_USUARIOS, H: http.HandlerFunc(ucontroller.Listar)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_USUARIOS_FORMNEW, H: http.HandlerFunc(ucontroller.FormNew)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_USUARIOS_CREAR, H: http.HandlerFunc(ucontroller.Crear)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_USUARIOS_GET, H: http.HandlerFunc(ucontroller.FormNew)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_USUARIOS_VER, H: http.HandlerFunc(ucontroller.Ver)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_USUARIOS_EDIT, H: http.HandlerFunc(ucontroller.EditarPerfil)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_USUARIOS_EDITSET, H: http.HandlerFunc(ucontroller.EditarPerfil)})

	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_ROLES, H: http.HandlerFunc(rcontroller.ListarRoles)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_ROLES_FORMNEW, H: http.HandlerFunc(rcontroller.FormNew)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_ROLES_CREAR, H: http.HandlerFunc(rcontroller.Crear)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_ROLES_GET, H: http.HandlerFunc(rcontroller.FormNew)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_ROLES_VER, H: http.HandlerFunc(rcontroller.Ver)})

	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_UNIDADES, H: http.HandlerFunc(uncontroller.Listar)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_UNIDADES_FORMNEW, H: http.HandlerFunc(uncontroller.FormNew)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_UNIDADES_CREAR, H: http.HandlerFunc(uncontroller.Crear)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_UNIDADES_GET, H: http.HandlerFunc(uncontroller.FormNew)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_UNIDADES_VER, H: http.HandlerFunc(uncontroller.Ver)})

	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_AVISOS, H: http.HandlerFunc(acontroller.Listar)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_AVISOS_FORMNEW, H: http.HandlerFunc(acontroller.FormNew)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_AVISOS_CREAR, H: http.HandlerFunc(acontroller.Crear)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_AVISOS_GET, H: http.HandlerFunc(acontroller.FormNew)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_AVISOS_VER, H: http.HandlerFunc(acontroller.Ver)})

	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_TICKETS, H: http.HandlerFunc(tcontroller.Listar)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_TICKETS_FORMNEW, H: http.HandlerFunc(tcontroller.FormNew)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_TICKETS_CREAR, H: http.HandlerFunc(tcontroller.Crear)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_TICKETS_VER, H: http.HandlerFunc(tcontroller.Ver)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_TICKETS_RESPONDER_GET, H: http.HandlerFunc(tcontroller.FormResponder)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_TICKETS_RESPONDER, H: http.HandlerFunc(tcontroller.ResponderTicket)})

	return ssr
}
