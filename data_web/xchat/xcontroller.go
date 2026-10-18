package xchat

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/dataauth/utilidad"
	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type Controller struct{ DB *sql.DB }

type chatMessageInput struct {
	ConversationID string `json:"chatConversationID"`
	RecipientID    string `json:"chatRecipientID"`
	Text           string `json:"chatText"`
}

func (c *Controller) Chat(w http.ResponseWriter, r *http.Request) {
	auth, err := utilidad.CtxValue(r.Context(), c.DB, "")
	if err != nil || auth == nil || auth.SessionKey == nil || auth.SessionKey.UsuarioID == "" {
		http.Error(w, "Sesión no autorizada", http.StatusUnauthorized)
		return
	}
	conversations, err := xrepo.ChatsByUser(r.Context(), c.DB, auth.SessionKey.UsuarioID)
	if err != nil {
		c.chatError(w, r, "No fue posible cargar tus conversaciones.")
		return
	}
	var body bytes.Buffer
	if err := ChatConversationList(conversations).Render(r.Context(), &body); err != nil {
		c.chatError(w, r, "No fue posible renderizar tus conversaciones.")
		return
	}
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#chat-conversation-list"), datastar.WithModeOuter())
	patchChatEmptyView(sse, r, false)
	patchUnreadBadges(sse, unreadCount(r.Context(), c.DB, auth.SessionKey.UsuarioID))
	_ = sse.ExecuteScript("document.getElementById('chat-empty-view')?.classList.remove('hidden');document.getElementById('chat-conversation-view')?.classList.add('hidden');document.getElementById('chat-users-view')?.classList.add('hidden');const title=document.getElementById('chat-dialog-title');if(title)title.textContent='Chat';")
}

func (c *Controller) ChatConversation(w http.ResponseWriter, r *http.Request) {
	auth, err := utilidad.CtxValue(r.Context(), c.DB, "")
	if err != nil || auth == nil || auth.SessionKey == nil || auth.SessionKey.UsuarioID == "" {
		http.Error(w, "Sesión no autorizada", http.StatusUnauthorized)
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("conversationid"))
	if id == "" {
		c.chatError(w, r, "No se recibió la conversación.")
		return
	}
	conversations, err := xrepo.ChatsByUser(r.Context(), c.DB, auth.SessionKey.UsuarioID)
	if err != nil {
		c.chatError(w, r, "No fue posible cargar tus conversaciones.")
		return
	}
	var conversation *model.ResponseChatConversacion
	for _, item := range conversations {
		if item != nil && item.ID == id {
			conversation = item
			break
		}
	}
	if conversation == nil {
		c.chatError(w, r, "La conversación no existe o no está disponible.")
		return
	}
	messages, err := xrepo.ChatMensajes(r.Context(), c.DB, id)
	if err != nil {
		c.chatError(w, r, "No fue posible cargar los mensajes.")
		return
	}
	recipientID := conversationRecipient(conversation, auth.SessionKey.UsuarioID)
	var body bytes.Buffer
	_ = ChatConversationView(conversation, messages, auth.SessionKey.UsuarioID, recipientID, conversation.Nombre).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	patchChatConversationList(sse, r, conversations)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#chat-conversation-view"), datastar.WithModeOuter())
	patchChatEmptyView(sse, r, true)
	patchUnreadBadges(sse, unreadCount(r.Context(), c.DB, auth.SessionKey.UsuarioID))
	_ = sse.MarshalAndPatchSignals(map[string]any{"chatConversationID": id, "chatRecipientID": recipientID, "chatRecipientName": conversation.Nombre, "chatText": "", "chatSending": false, "chatScrollNonce": time.Now().UnixNano()})
	_ = sse.ExecuteScript("document.getElementById('chat-users-view')?.classList.add('hidden');document.getElementById('chat-conversation-view')?.classList.remove('hidden');" + chatScrollToBottomScript() + "const title=document.getElementById('chat-dialog-title');if(title)title.textContent=" + strconv.Quote(conversation.Nombre) + ";")
}

func (c *Controller) ChatMessages(w http.ResponseWriter, r *http.Request) {
	auth, err := utilidad.CtxValue(r.Context(), c.DB, "")
	if err != nil || auth == nil || auth.SessionKey == nil || auth.SessionKey.UsuarioID == "" {
		http.Error(w, "Sesión no autorizada", http.StatusUnauthorized)
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("conversationid"))
	if id == "" {
		c.chatError(w, r, "No se recibió la conversación.")
		return
	}
	conversations, err := xrepo.ChatsByUser(r.Context(), c.DB, auth.SessionKey.UsuarioID)
	if err != nil {
		c.chatError(w, r, "No fue posible cargar tus conversaciones.")
		return
	}
	var conversation *model.ResponseChatConversacion
	for _, item := range conversations {
		if item != nil && item.ID == id {
			conversation = item
			break
		}
	}
	if conversation == nil {
		c.chatError(w, r, "La conversación no existe o no está disponible.")
		return
	}
	messages, err := xrepo.ChatMensajes(r.Context(), c.DB, id)
	if err != nil {
		c.chatError(w, r, "No fue posible cargar los mensajes.")
		return
	}
	var body bytes.Buffer
	_ = ChatMessages(messages, auth.SessionKey.UsuarioID).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	patchChatConversationList(sse, r, conversations)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#chat-messages"), datastar.WithModeOuter())
	patchUnreadBadges(sse, unreadCount(r.Context(), c.DB, auth.SessionKey.UsuarioID))
	_ = sse.MarshalAndPatchSignals(map[string]any{"chatScrollNonce": time.Now().UnixNano()})
	_ = sse.ExecuteScript(chatScrollToBottomScript())
}

func (c *Controller) ChatUsers(w http.ResponseWriter, r *http.Request) {
	auth, err := utilidad.CtxValue(r.Context(), c.DB, "")
	if err != nil || auth == nil || auth.SessionKey == nil {
		http.Error(w, "Sesión no autorizada", http.StatusUnauthorized)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	var nombre *string
	if query != "" {
		nombre = &query
	}
	users, err := xrepo.Usuarios(r.Context(), c.DB, model.QueryUsuarios{Pagina: 1, Size: 100, Nombre: nombre})
	if err != nil {
		c.chatError(w, r, "No fue posible cargar los usuarios.")
		return
	}
	filtered := make([]*model.Usuario, 0, len(users))
	for _, user := range users {
		if user != nil && user.ID != auth.SessionKey.UsuarioID {
			filtered = append(filtered, user)
		}
	}
	var body bytes.Buffer
	_ = ChatUserList(filtered).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#chat-users"), datastar.WithModeOuter())
	patchChatEmptyView(sse, r, true)
}

func (c *Controller) ChatCompose(w http.ResponseWriter, r *http.Request) {
	auth, err := utilidad.CtxValue(r.Context(), c.DB, "")
	if err != nil || auth == nil || auth.SessionKey == nil {
		http.Error(w, "Sesión no autorizada", http.StatusUnauthorized)
		return
	}
	recipientID := strings.TrimSpace(r.URL.Query().Get("userid"))
	recipientName := strings.TrimSpace(r.URL.Query().Get("name"))
	if recipientID == "" || recipientName == "" || recipientID == auth.SessionKey.UsuarioID {
		c.chatError(w, r, "El destinatario seleccionado no es válido.")
		return
	}
	var body bytes.Buffer
	_ = ChatConversationView(nil, nil, auth.SessionKey.UsuarioID, recipientID, recipientName).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#chat-conversation-view"), datastar.WithModeOuter())
	patchChatEmptyView(sse, r, true)
	_ = sse.MarshalAndPatchSignals(map[string]any{"chatConversationID": "", "chatRecipientID": recipientID, "chatRecipientName": recipientName, "chatText": "", "chatSending": false})
	_ = sse.ExecuteScript("document.getElementById('chat-users-view')?.classList.add('hidden');document.getElementById('chat-conversation-view')?.classList.remove('hidden');" + chatScrollToBottomScript() + "const title=document.getElementById('chat-dialog-title');if(title)title.textContent=" + strconv.Quote(recipientName) + ";")
}

func (c *Controller) ChatSend(w http.ResponseWriter, r *http.Request) {
	auth, err := utilidad.CtxValue(r.Context(), c.DB, "")
	if err != nil || auth == nil || auth.SessionKey == nil {
		http.Error(w, "Sesión no autorizada", http.StatusUnauthorized)
		return
	}
	var signals chatMessageInput
	if err := datastar.ReadSignals(r, &signals); err != nil {
		_ = r.ParseForm()
		signals.ConversationID = r.FormValue("chatConversationID")
		signals.RecipientID = r.FormValue("chatRecipientID")
		signals.Text = r.FormValue("chatText")
	}
	if signals.RecipientID == "" {
		signals.RecipientID = strings.TrimSpace(r.URL.Query().Get("recipientid"))
	}
	signals.Text = strings.TrimSpace(signals.Text)
	if signals.RecipientID == "" || signals.Text == "" {
		c.chatError(w, r, "Selecciona un destinatario y escribe un mensaje.")
		return
	}
	message, err := xrepo.ChatEnviarMensaje(r.Context(), c.DB, model.ChatEnviarMensajeInput{SenderID: auth.SessionKey.UsuarioID, DestinatorID: signals.RecipientID, Tipo: model.ChatMensajeTipoText, Texto: signals.Text})
	if err != nil || message == nil {
		if err == nil {
			err = fmt.Errorf("no fue posible enviar el mensaje")
		}
		c.chatError(w, r, err.Error())
		return
	}
	conversations, err := xrepo.ChatsByUser(r.Context(), c.DB, auth.SessionKey.UsuarioID)
	if err != nil {
		c.chatError(w, r, "El mensaje se envió, pero no se pudo actualizar la conversación.")
		return
	}
	var conversation *model.ResponseChatConversacion
	for _, item := range conversations {
		if item != nil && item.ID == message.ConversacionID {
			conversation = item
			break
		}
	}
	messages, err := xrepo.ChatMensajes(r.Context(), c.DB, message.ConversacionID)
	if err != nil || conversation == nil {
		c.chatError(w, r, "El mensaje se envió, pero no se pudo actualizar la conversación.")
		return
	}
	var body bytes.Buffer
	_ = ChatMessages(messages, auth.SessionKey.UsuarioID).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	patchChatConversationList(sse, r, conversations)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#chat-messages"), datastar.WithModeOuter())
	patchChatEmptyView(sse, r, true)
	patchUnreadBadges(sse, unreadCount(r.Context(), c.DB, auth.SessionKey.UsuarioID))
	_ = sse.MarshalAndPatchSignals(map[string]any{"chatSending": false, "chatText": "", "chatConversationID": message.ConversacionID, "chatScrollNonce": time.Now().UnixNano()})
	_ = sse.ExecuteScript("const input=document.getElementById('chat-message-input');if(input)input.value='';" + chatScrollToBottomScript())
}

func patchChatEmptyView(sse *datastar.ServerSentEventGenerator, r *http.Request, hidden bool) {
	var body bytes.Buffer
	if err := ChatEmptyView(hidden).Render(r.Context(), &body); err == nil {
		_ = sse.PatchElements(body.String(), datastar.WithSelector("#chat-empty-view"), datastar.WithModeOuter())
	}
}

func patchChatConversationList(sse *datastar.ServerSentEventGenerator, r *http.Request, conversations []*model.ResponseChatConversacion) {
	var body bytes.Buffer
	if err := ChatConversationList(conversations).Render(r.Context(), &body); err == nil {
		_ = sse.PatchElements(body.String(), datastar.WithSelector("#chat-conversation-list"), datastar.WithModeOuter())
	}
}

func chatScrollToBottomScript() string {
	return "const scrollChat=()=>{const messages=document.getElementById('chat-messages');if(messages)messages.scrollTo({top:messages.scrollHeight,left:0,behavior:'auto'});};scrollChat();requestAnimationFrame(scrollChat);setTimeout(scrollChat,50);setTimeout(scrollChat,150);setTimeout(scrollChat,300);setTimeout(scrollChat,600);setTimeout(scrollChat,1000);"
}

func chatMessagesScrollEffect() string {
	return "const scrollNonce=$chatScrollNonce;const scrollChat=()=>{el.scrollTo({top:el.scrollHeight,left:0,behavior:'auto'});};scrollChat();requestAnimationFrame(scrollChat);setTimeout(scrollChat,50);setTimeout(scrollChat,150);setTimeout(scrollChat,300);setTimeout(scrollChat,600);setTimeout(scrollChat,1000);"
}

func (c *Controller) chatError(w http.ResponseWriter, r *http.Request, message string) {
	sse := datastar.NewSSE(w, r)
	var body bytes.Buffer
	_ = ChatError(message).Render(r.Context(), &body)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#chat-error"), datastar.WithModeOuter())
	_ = sse.MarshalAndPatchSignals(map[string]bool{"chatSending": false})
}

func unreadCount(ctx context.Context, db *sql.DB, userID string) int32 {
	count, err := xrepo.ChatsNoLeidos(ctx, db, userID)
	if err != nil || count < 0 {
		return 0
	}
	return count
}

func patchUnreadBadges(sse *datastar.ServerSentEventGenerator, count int32) {
	_ = sse.PatchElements(unreadBadgeHTML("chat-unread-badge-trigger", count), datastar.WithSelector("#chat-unread-badge-trigger"), datastar.WithModeOuter())
	_ = sse.PatchElements(unreadBadgeHTML("chat-unread-badge-menu", count), datastar.WithSelector("#chat-unread-badge-menu"), datastar.WithModeOuter())
}

func unreadBadgeHTML(id string, count int32) string {
	class := UnreadBadgeClass(count)
	return `<span id="` + id + `" data-chat-unread class="` + class + `">` + UnreadString(count) + `</span>`
}

func UnreadString(count int32) string {
	if count > 99 {
		return "99+"
	}
	return strconv.Itoa(int(count))
}

func UnreadBadgeClass(count int32) string {
	if count <= 0 {
		return "hidden rounded-full bg-primary px-1.5 py-0.5 text-[10px] font-semibold leading-none text-primary-foreground"
	}
	return "rounded-full bg-primary px-1.5 py-0.5 text-[10px] font-semibold leading-none text-primary-foreground"
}

func chatDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Local().Format("02/01/2006 15:04")
}
func chatInitials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "U"
	}
	if len(parts) == 1 {
		return strings.ToUpper(string([]rune(parts[0])[0]))
	}
	return strings.ToUpper(string([]rune(parts[0])[0]) + string([]rune(parts[len(parts)-1])[0]))
}
func chatPathURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "/") {
		return value
	}
	return "/" + strings.TrimPrefix(value, "./")
}
func chatConversationPhotoURL(c *model.ResponseChatConversacion) string {
	if c == nil || c.FotoURL == nil {
		return ""
	}
	return chatPathURL(*c.FotoURL)
}
func chatUserName(u *model.Usuario) string {
	if u == nil {
		return "Usuario"
	}
	return strings.TrimSpace(u.Nombres + " " + u.Apellido1)
}
func chatUserPhotoURL(u *model.Usuario) string {
	if u == nil || u.FotoURL == nil {
		return ""
	}
	return chatPathURL(*u.FotoURL)
}
func chatMessagesChronological(messages []*model.ChatMensaje) []*model.ChatMensaje {
	result := make([]*model.ChatMensaje, len(messages))
	for i, m := range messages {
		result[len(messages)-i-1] = m
	}
	return result
}
func chatMessageClass(message *model.ChatMensaje, currentUserID string) string {
	if message != nil && message.SenderID == currentUserID {
		return "ml-auto max-w-[85%] rounded-2xl rounded-br-sm bg-primary px-3 py-2 text-primary-foreground"
	}
	return "mr-auto max-w-[85%] rounded-2xl rounded-bl-sm bg-background px-3 py-2 text-foreground shadow-sm"
}
func conversationRecipient(c *model.ResponseChatConversacion, current string) string {
	if c == nil {
		return ""
	}
	for _, m := range c.Miembros {
		if m != nil && m.UsuarioID != current {
			return m.UsuarioID
		}
	}
	return ""
}

func ChatConversationClick(id string) string {
	return "$chatConversationID=" + strconv.Quote(id) + ";$chatSending=false;$chatText='';@get('" + ADMIN_CHAT_CONVERSATION + "?conversationid='+encodeURIComponent($chatConversationID))"
}
func ChatUserClick(u *model.Usuario) string {
	if u == nil {
		return ""
	}
	name := chatUserName(u)
	return "$chatRecipientID=" + strconv.Quote(u.ID) + ";$chatRecipientName=" + strconv.Quote(name) + ";$chatConversationID='';$chatText='';$chatSending=false;@get('" + ADMIN_CHAT_COMPOSE + "?userid='+encodeURIComponent($chatRecipientID)+'&name='+encodeURIComponent($chatRecipientName))"
}
func ChatNewScript() string {
	return "$chatConversationID='';$chatRecipientID='';$chatRecipientName='';$chatText='';$chatSending=false;document.getElementById('chat-empty-view')?.classList.add('hidden');document.getElementById('chat-conversation-view')?.classList.add('hidden');document.getElementById('chat-users-view')?.classList.remove('hidden');document.getElementById('chat-back-button')?.classList.remove('hidden');const title=document.getElementById('chat-dialog-title');if(title)title.textContent='Nuevo chat';@get('" + ADMIN_CHAT_USERS + "')"
}
func ChatBackScript() string {
	return "$chatConversationID='';$chatRecipientID='';$chatRecipientName='';$chatText='';$chatSending=false;document.getElementById('chat-users-view')?.classList.add('hidden');document.getElementById('chat-conversation-view')?.classList.add('hidden');document.getElementById('chat-empty-view')?.classList.remove('hidden');document.getElementById('chat-back-button')?.classList.add('hidden');const title=document.getElementById('chat-dialog-title');if(title)title.textContent='Chat';@get('" + ADMIN_CHAT + "')"
}
func ChatNotificationHandler() string {
	return "const data=evt.detail;if(data?.tipo==='chat'){const id=String(data.datos?.conversation_id||'');if($chatConversationID&&id===String($chatConversationID)){@get('" + ADMIN_CHAT_MESSAGES + "?conversationid='+encodeURIComponent(String($chatConversationID)))}else{@get('" + ADMIN_CHAT + "')}}"
}
func ChatSearchRequestScript() string {
	return "const input=document.getElementById('chat-user-search');const value=input?.value||'';$chatSearch=value;@get('" + ADMIN_CHAT_USERS + "?q='+encodeURIComponent(value))"
}
func ChatTriggerAttributes() templ.Attributes {
	return templ.Attributes{"tabindex": "0", "type": "button", "data-slot": "dialog-trigger", "data-base-ui-click-trigger": true, "data-templ-controls": "chat-dialog", "aria-haspopup": "dialog", "aria-expanded": "false", "data-on:click": "$chatConversationID='';$chatText='';$chatSending=false;@get('" + ADMIN_CHAT + "')"}
}
