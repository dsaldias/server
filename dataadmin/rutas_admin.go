package dataadmin

import (
	"database/sql"
	"embed"
	"io/fs"
	"net/http"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/dataadmin/admin/avisos"
	"github.com/dsaldias/server/dataadmin/admin/mainlayout"
	xlogin "github.com/dsaldias/server/dataadmin/admin/mainlayout/login"
	"github.com/dsaldias/server/dataadmin/admin/roles"
	"github.com/dsaldias/server/dataadmin/admin/ticket"
	"github.com/dsaldias/server/dataadmin/admin/unidades"
	"github.com/dsaldias/server/dataadmin/admin/usuarios"
	"github.com/dsaldias/server/dataauth/utils"
)

//go:embed assets/*
var Assets embed.FS

var (
	WEB_PATH_BASE  = "/webx/"
	WEB_PATH_LOGIN = "/webx/login"
	WEB_PATH_MAIN  = "/webx/main"

	WEB_ADMIN_PATH_SET_COOKIE_UNIDAD = "/webx/setcookie/data"
	WEB_ADMIN_PATH_USUARIOS          = "/webx/usuarios"
	WEB_ADMIN_PATH_USUARIOS_FORMNEW  = "/webx/usuarios/formnew"
	WEB_ADMIN_PATH_USUARIOS_CREAR    = "/webx/usuarios/crear"
	WEB_ADMIN_PATH_USUARIOS_GET      = "/webx/usuarios/{id}/get"
	WEB_ADMIN_PATH_USUARIOS_VER      = "/webx/usuarios/{id}/ver"

	WEB_ADMIN_PATH_ROLES         = "/webx/roles"
	WEB_ADMIN_PATH_ROLES_FORMNEW = "/webx/roles/formnew"
	WEB_ADMIN_PATH_ROLES_CREAR   = "/webx/roles/crear"
	WEB_ADMIN_PATH_ROLES_GET     = "/webx/roles/{id}/get"
	WEB_ADMIN_PATH_ROLES_VER     = "/webx/roles/{id}/ver"

	WEB_ADMIN_PATH_UNIDADES         = "/webx/unidades"
	WEB_ADMIN_PATH_UNIDADES_FORMNEW = "/webx/unidades/formnew"
	WEB_ADMIN_PATH_UNIDADES_CREAR   = "/webx/unidades/crear"
	WEB_ADMIN_PATH_UNIDADES_GET     = "/webx/unidades/{id}/get"
	WEB_ADMIN_PATH_UNIDADES_VER     = "/webx/unidades/{id}/ver"

	WEB_ADMIN_PATH_AVISOS         = "/webx/avisos"
	WEB_ADMIN_PATH_AVISOS_FORMNEW = "/webx/avisos/formnew"
	WEB_ADMIN_PATH_AVISOS_CREAR   = "/webx/avisos/crear"
	WEB_ADMIN_PATH_AVISOS_GET     = "/webx/avisos/{id}/get"
	WEB_ADMIN_PATH_AVISOS_VER     = "/webx/avisos/{id}/ver"

	WEB_ADMIN_PATH_TICKETS               = "/webx/tickets"
	WEB_ADMIN_PATH_TICKETS_FORMNEW       = "/webx/tickets/formnew"
	WEB_ADMIN_PATH_TICKETS_CREAR         = "/webx/tickets/crear"
	WEB_ADMIN_PATH_TICKETS_VER           = "/webx/tickets/{id}/ver"
	WEB_ADMIN_PATH_TICKETS_RESPONDER_GET = "/webx/tickets/{id}/get"
	WEB_ADMIN_PATH_TICKETS_RESPONDER     = "/webx/tickets/responder"
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

	ssr = append(ssr, &utils.Handlers2{Path: WEB_PATH_LOGIN, H: cont_login.Login()})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_PATH_BASE, H: templ.Handler(xlogin.Inicio())})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_PATH_MAIN, H: http.HandlerFunc(cont_main.MainLayout)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_SET_COOKIE_UNIDAD, H: http.HandlerFunc(umain.SetCookieUnidadId)})

	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_USUARIOS, H: http.HandlerFunc(ucontroller.Listar)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_USUARIOS_FORMNEW, H: http.HandlerFunc(ucontroller.FormNew)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_USUARIOS_CREAR, H: http.HandlerFunc(ucontroller.Crear)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_USUARIOS_GET, H: http.HandlerFunc(ucontroller.FormNew)})
	ssr = append(ssr, &utils.Handlers2{Path: WEB_ADMIN_PATH_USUARIOS_VER, H: http.HandlerFunc(ucontroller.Ver)})

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
