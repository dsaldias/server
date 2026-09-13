package repo

import (
	"context"
	"database/sql"

	"github.com/dsaldias/server/dataauth/archivos"
	"github.com/dsaldias/server/dataauth/avisos"
	"github.com/dsaldias/server/dataauth/chat"
	"github.com/dsaldias/server/dataauth/dashboard"
	"github.com/dsaldias/server/dataauth/login"
	"github.com/dsaldias/server/dataauth/menus"
	"github.com/dsaldias/server/dataauth/permisos"
	"github.com/dsaldias/server/dataauth/roles"
	"github.com/dsaldias/server/dataauth/ticketss"
	"github.com/dsaldias/server/dataauth/unidades"
	"github.com/dsaldias/server/dataauth/usuarios"
	"github.com/dsaldias/server/dataauth/utils"
	"github.com/dsaldias/server/dataauth/xnotificaciones"
	"github.com/dsaldias/server/graph_auth/model"
)

// Login is the resolver for the login field.
func Login(ctx context.Context, db *sql.DB, input model.NewLogin) (*model.ResponseLogin, error) {
	return login.Login(ctx, db, input, false)
}

// LoginV2 is the resolver for the login_v2 field.
func LoginV2(ctx context.Context, db *sql.DB, input model.NewLogin2) (*model.ResponseLogin, error) {
	return login.Login2(ctx, db, input)
}

// CreateRol is the resolver for the create_rol field.
func CreateRol(ctx context.Context, db *sql.DB, input model.NewRol) (*model.Rol, error) {
	_, err := utils.CtxValue(ctx, db, "create_rol")
	if err != nil {
		return nil, err
	}
	return roles.Crear(db, input)
}

// UpdateRol is the resolver for the update_rol field.
func UpdateRol(ctx context.Context, db *sql.DB, input model.UpdateRol) (*model.Rol, error) {
	_, err := utils.CtxValue(ctx, db, "update_rol")
	if err != nil {
		return nil, err
	}
	return roles.Actualizar(db, input)
}

// CreateUsuario is the resolver for the create_usuario field.
func CreateUsuario(ctx context.Context, db *sql.DB, input model.NewUsuario) (*model.Usuario, error) {
	_, err := utils.CtxValue(ctx, db, "create_usuario")
	if err != nil {
		return nil, err
	}
	return usuarios.Crear(db, input, nil)
}

// UpdateUsuario is the resolver for the update_usuario field.
func UpdateUsuario(ctx context.Context, db *sql.DB, input model.UpdateUsuario) (*model.Usuario, error) {
	_, err := utils.CtxValue(ctx, db, "update_usuario")
	if err != nil {
		return nil, err
	}
	return usuarios.Actualizar(db, input)
}

// UpdatePerfil is the resolver for the update_perfil field.
func UpdatePerfil(ctx context.Context, db *sql.DB, input model.UpdatePerfil) (*model.Usuario, error) {
	_, err := utils.CtxValue(ctx, db, "update_perfil")
	if err != nil {
		return nil, err
	}
	return usuarios.UpdatePerfil(db, input)
}

// CreateUnidad is the resolver for the create_unidad field.
func CreateUnidad(ctx context.Context, db *sql.DB, input model.NewUnidad) (*model.Unidad, error) {
	_, err := utils.CtxValue(ctx, db, "create_unidad")
	if err != nil {
		return nil, err
	}
	return unidades.Crear(db, input)
}

// UpdateUnidad is the resolver for the update_unidad field.
func UpdateUnidad(ctx context.Context, db *sql.DB, input model.UpdUnidad) (*model.Unidad, error) {
	_, err := utils.CtxValue(ctx, db, "update_unidad")
	if err != nil {
		return nil, err
	}
	return unidades.Actualizar(db, input)
}

// CreateOauth is the resolver for the createOauth field.
func CreateOauth(ctx context.Context, db *sql.DB, input model.NewUsuarioOauth) (*model.Usuario, error) {
	return usuarios.CrearOauth(db, input, false)
}

// EnviarNotificacion is the resolver for the enviar_notificacion field.
func EnviarNotificacion(ctx context.Context, db *sql.DB, titulo string) (bool, error) {
	return xnotificaciones.EnviarNotificacion(ctx, titulo, nil)
}

// CrearNotificacion is the resolver for the crear_notificacion field.
func CrearNotificacion(ctx context.Context, db *sql.DB, input model.NewNotificacion) (*model.Notificacion, error) {
	tok, err := utils.CtxValue(ctx, db, "crear_notificacion")
	if err != nil {
		return nil, err
	}
	userid := tok.SessionKey.UsuarioID
	return avisos.Crear(ctx, db, input, userid)
}

// UpdateNotificacion is the resolver for the update_notificacion field.
func UpdateNotificacion(ctx context.Context, db *sql.DB, input model.UpdNotificacion) (*model.Notificacion, error) {
	tok, err := utils.CtxValue(ctx, db, "update_notificacion")
	if err != nil {
		return nil, err
	}
	userid := tok.SessionKey.UsuarioID
	return avisos.Actualizar(db, input, userid)
}

// CreateTicket is the resolver for the create_ticket field.
func CreateTicket(ctx context.Context, db *sql.DB, input model.NewTicket) (*model.Ticket, error) {
	tok, err := utils.CtxValue(ctx, db, "create_ticket")
	if err != nil {
		return nil, err
	}
	userid := tok.SessionKey.UsuarioID
	return ticketss.Crear(ctx, db, input, userid)
}

// UpdateTicket is the resolver for the update_ticket field.
func UpdateTicket(ctx context.Context, db *sql.DB, input model.NewTicketRespuesta) (*model.Ticket, error) {
	tok, err := utils.CtxValue(ctx, db, "update_ticket")
	if err != nil {
		return nil, err
	}
	userid := tok.SessionKey.UsuarioID
	return ticketss.Update(ctx, db, input, userid)
}

// CerrarTicket is the resolver for the cerrar_ticket field.
func CerrarTicket(ctx context.Context, db *sql.DB, id string) (*model.Ticket, error) {
	_, err := utils.CtxValue(ctx, db, "cerrar_ticket")
	if err != nil {
		return nil, err
	}
	return ticketss.Cerrar(ctx, db, id)
}

// ChatEnviarMensaje is the resolver for the chat_enviar_mensaje field.
func ChatEnviarMensaje(ctx context.Context, db *sql.DB, input model.ChatEnviarMensajeInput) (*model.ChatMensaje, error) {
	return chat.EnviarChatMensaje(ctx, db, input)
}

// Me is the resolver for the me field.
func Me(ctx context.Context, db *sql.DB, input model.InputMe) (*model.ResponseMe, error) {
	tok, err := utils.CtxValue(ctx, db, "")
	if err != nil {
		return nil, err
	}
	userid := tok.SessionKey.UsuarioID
	return usuarios.GetMe(db, input, userid)
}

// Roles is the resolver for the roles field.
func Roles(ctx context.Context, db *sql.DB) ([]*model.ResponseRoles, error) {
	_, err := utils.CtxValue(ctx, db, "roles")
	if err != nil {
		return nil, err
	}
	return roles.GetRoles(db)
}

// Permisos is the resolver for the permisos field.
func Permisos(ctx context.Context, db *sql.DB) ([]*model.Permiso, error) {
	_, err := utils.CtxValue(ctx, db, "permisos")
	if err != nil {
		return nil, err
	}
	return permisos.GetPermisos(db)
}

// Usuarios is the resolver for the usuarios field.
func Usuarios(ctx context.Context, db *sql.DB, query model.QueryUsuarios) ([]*model.Usuario, error) {
	_, err := utils.CtxValue(ctx, db, "usuarios")
	if err != nil {
		return nil, err
	}
	return usuarios.GetUsuarios(db, query)
}

// UsuariosConectados is the resolver for the usuarios_conectados field.
func UsuariosConectados(ctx context.Context, db *sql.DB) ([]*model.Usuario, error) {
	return usuarios.GetUsuariosConectados(db)
}

// UsuarioByID is the resolver for the usuario_by_id field.
func UsuarioByID(ctx context.Context, db *sql.DB, id string) (*model.ResponseUsuario, error) {
	_, err := utils.CtxValue(ctx, db, "usuario_by_id")
	if err != nil {
		return nil, err
	}
	return usuarios.GetBy(db, id)
}

// RolByID is the resolver for the rol_by_id field.
func RolByID(ctx context.Context, db *sql.DB, id string) (*model.Rol, error) {
	_, err := utils.CtxValue(ctx, db, "rol_by_id")
	if err != nil {
		return nil, err
	}
	return roles.GetRolById(db, id)
}

// Menus is the resolver for the menus field.
func Menus(ctx context.Context, db *sql.DB) ([]*model.Menus, error) {
	_, err := utils.CtxValue(ctx, db, "menus")
	if err != nil {
		return nil, err
	}
	return menus.Listar(db)
}

// Unidades is the resolver for the unidades field.
func Unidades(ctx context.Context, db *sql.DB) ([]*model.Unidad, error) {
	_, err := utils.CtxValue(ctx, db, "unidades")
	if err != nil {
		return nil, err
	}
	return unidades.Listar(db)
}

func UnidadByID(ctx context.Context, db *sql.DB, id string) (*model.Unidad, error) {
	_, err := utils.CtxValue(ctx, db, "")
	if err != nil {
		return nil, err
	}
	return unidades.GetById(db, id)
}

// GetImagen is the resolver for the get_imagen field.
func GetImagen(ctx context.Context, db *sql.DB, url string) (string, error) {
	return archivos.GetImagen(url)
}

// ConexionesWs is the resolver for the conexiones_ws field.
func ConexionesWs(ctx context.Context, db *sql.DB) (string, error) {
	return xnotificaciones.VerConexiones()
}

// Notificaciones is the resolver for the notificaciones field.
func Notificaciones(ctx context.Context, db *sql.DB) ([]*model.Notificacion, error) {
	return avisos.GetNotificacionesActivas(db)
}

func Notificacion(ctx context.Context, db *sql.DB, id string) (*model.Notificacion, error) {
	return avisos.Get(db, id)
}

// Reporte1 is the resolver for the reporte1 field.
func Reporte1(ctx context.Context, db *sql.DB) ([]*model.ResponseReporte1, error) {
	return dashboard.Reporte1(db)
}

// Reporte2 is the resolver for the reporte2 field.
func Reporte2(ctx context.Context, db *sql.DB) ([]*model.ResponseReporte2, error) {
	return dashboard.Reporte2(db)
}

// Reporte2b is the resolver for the reporte2b field.
func Reporte2b(ctx context.Context, db *sql.DB) ([]*model.ResponseReporte2b, error) {
	return dashboard.Reporte2b(db)
}

// AllTickets is the resolver for the all_tickets field.
func AllTickets(ctx context.Context, db *sql.DB, q model.QueryTickets) ([]*model.RespTickets, error) {
	_, err := utils.CtxValue(ctx, db, "all_tickets")
	if err != nil {
		return nil, err
	}
	return ticketss.AllTickets(db, q)
}

// MisTickets is the resolver for the mis_tickets field.
func MisTickets(ctx context.Context, db *sql.DB) ([]*model.RespTickets, error) {
	tok, err := utils.CtxValue(ctx, db, "mis_tickets")
	if err != nil {
		return nil, err
	}
	userid := tok.SessionKey.UsuarioID
	return ticketss.MisTickets(db, userid)
}

// VerTicket is the resolver for the ver_ticket field.
func VerTicket(ctx context.Context, db *sql.DB, id string) (*model.Ticket, error) {
	_, err := utils.CtxValue(ctx, db, "ver_ticket")
	if err != nil {
		return nil, err
	}
	return ticketss.Get(ctx, db, id)
}

// ChatsByUser is the resolver for the chats_by_user field.
func ChatsByUser(ctx context.Context, db *sql.DB, userID string) ([]*model.ResponseChatConversacion, error) {
	return chat.ConversacionesByUser(db, userID)
}

// ChatMensajes is the resolver for the chat_mensajes field.
func ChatMensajes(ctx context.Context, db *sql.DB, conversacionID string) ([]*model.ChatMensaje, error) {
	tok, err := utils.CtxValue(ctx, db, "")
	if err != nil {
		return nil, err
	}
	userid := tok.SessionKey.UsuarioID
	return chat.MensajesByChat(db, conversacionID, userid)
}

// ChatsNoLeidos is the resolver for the chats_no_leidos field.
func ChatsNoLeidos(ctx context.Context, db *sql.DB, userID string) (int32, error) {
	return chat.MensajesNoLeidos(db, userID)
}

// NotificacionesSubs is the resolver for the notificaciones_subs field.
func NotificacionesSubs(ctx context.Context, db *sql.DB) (<-chan *model.XNotificacion, error) {
	userid := utils.CtxUserIDWs(ctx, db, "")
	return xnotificaciones.NotificacionesSubs(ctx, userid)
}
