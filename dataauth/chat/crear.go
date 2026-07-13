package chat

import (
	"context"
	"database/sql"
	"strconv"

	rbacconversaciones "github.com/dsaldias/server/dataauth/chat/rbac_conversaciones"
	rbacmensajes "github.com/dsaldias/server/dataauth/chat/rbac_mensajes"
	rbacstatus "github.com/dsaldias/server/dataauth/chat/rbac_status"
	"github.com/dsaldias/server/dataauth/xnotificaciones"
	"github.com/dsaldias/server/graph_auth/model"
)

func EnviarChatMensaje(ctx context.Context, db *sql.DB, input model.ChatEnviarMensajeInput) (*model.ChatMensaje, error) {

	conversacion_id, err := rbacconversaciones.BuscarConversacionPrivada(db, input.SenderID, input.DestinatorID)
	if err != nil {
		return nil, err
	}

	if conversacion_id == -1 {
		tx, err := db.Begin()
		if err != nil {
			return nil, err
		}

		defer tx.Rollback()

		conver := model.NewChatConversacion{
			Tipo:         "privado",
			Nombre:       "",
			CreatedBy:    input.SenderID,
			DestinatorID: input.DestinatorID,
		}
		conversacion_id, err = rbacconversaciones.CrearConMiembros(tx, conver)
		if err != nil {
			return nil, err
		}

		err = tx.Commit()
		if err != nil {
			return nil, err
		}

	}

	conversacionid := strconv.FormatInt(conversacion_id, 10)

	msg := model.NewChatMensaje{
		SenderID:       input.SenderID,
		ConversacionID: conversacionid,
		Tipo:           input.Tipo,
		Texto:          input.Texto,
	}
	men, err := rbacmensajes.Crear(db, msg)
	if err != nil {
		return nil, err
	}

	tatus := model.NewChatMensajeStatus{
		MessageID: men.ID,
		UsuarioID: input.SenderID,
		Status:    "sent",
	}
	_, err = rbacstatus.Crear(db, tatus)
	if err != nil {
		return nil, err
	}

	tipo := "chat"
	mapa := map[string]*string{}
	mapa["sender_id"] = &input.SenderID
	mapa["destinator_id"] = &input.DestinatorID
	mapa["conversation_id"] = &conversacionid

	d := xnotificaciones.DataNotify{Tipo: &tipo, Datos: mapa}
	xnotificaciones.EnviarNotificacionAUsuario(ctx, input.DestinatorID, "chat mensaje", &d)
	return men, nil
}
