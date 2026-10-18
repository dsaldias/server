package xchat

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

const (
	ADMIN_CHAT              = "/adminx/chat"
	ADMIN_CHAT_USERS        = ADMIN_CHAT + "/usuarioshtml"
	ADMIN_CHAT_CONVERSATION = ADMIN_CHAT + "/conversacionhtml"
	ADMIN_CHAT_MESSAGES     = ADMIN_CHAT + "/mensajeshtml"
	ADMIN_CHAT_COMPOSE      = ADMIN_CHAT + "/componerhtml"
	ADMIN_CHAT_SEND         = ADMIN_CHAT + "/enviar"
)

func Rutas(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{DB: db}
	return []*utilidad.Handlers2{
		{Path: ADMIN_CHAT, H: http.HandlerFunc(controller.Chat)},
		{Path: ADMIN_CHAT_USERS, H: http.HandlerFunc(controller.ChatUsers)},
		{Path: ADMIN_CHAT_CONVERSATION, H: http.HandlerFunc(controller.ChatConversation)},
		{Path: ADMIN_CHAT_MESSAGES, H: http.HandlerFunc(controller.ChatMessages)},
		{Path: ADMIN_CHAT_COMPOSE, H: http.HandlerFunc(controller.ChatCompose)},
		{Path: ADMIN_CHAT_SEND, H: http.HandlerFunc(controller.ChatSend)},
	}
}
