package chat

import (
	"database/sql"

	rbacconversaciones "github.com/dsaldias/server/dataauth/chat/rbac_conversaciones"
	rbacconversacionesmiembros "github.com/dsaldias/server/dataauth/chat/rbac_conversaciones_miembros"
	rbacmensajes "github.com/dsaldias/server/dataauth/chat/rbac_mensajes"
	rbacstatus "github.com/dsaldias/server/dataauth/chat/rbac_status"
	"github.com/dsaldias/server/graph_auth/model"
)

func ConversacionesByUser(db *sql.DB, user_id string) ([]*model.ResponseChatConversacion, error) {
	return rbacconversaciones.ByUser(db, user_id)
}

func MensajesNoLeidos(db *sql.DB, user_id string) (int32, error) {
	return rbacstatus.NoLeidos(db, user_id)
}

func MensajesByChat(db *sql.DB, conversacion_id string) ([]*model.ChatMensaje, error) {
	miembros, err := rbacconversacionesmiembros.ByConversacion(db, conversacion_id)
	if err != nil {
		return nil, err
	}
	return rbacmensajes.ByChat(db, conversacion_id, miembros)
}
