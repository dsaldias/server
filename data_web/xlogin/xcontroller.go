package xlogin

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/data_web/components/combobox"
	"github.com/dsaldias/server/data_web/xutility"
	"github.com/dsaldias/server/dataauth/utilidad"
	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type Controller struct {
	DB *sql.DB
}

type InicioData struct {
	Authenticated bool
	Me            *model.ResponseMe
	ChatUnread    int32
	UnidadID      string
	RolID         string
	Content       templ.Component
	ActivePath    string
}

type contextInput struct {
	Context string `json:"context"`
}

type loginSubmitInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Context  string `json:"context"`
}

type perfilSubmitInput struct {
	Nombres   string `json:"nombres"`
	Apellido1 string `json:"apellido1"`
	Apellido2 string `json:"apellido2"`
	Documento string `json:"documento"`
	Celular   string `json:"celular"`
	Correo    string `json:"correo"`
	Sexo      string `json:"sexo"`
	Direccion string `json:"direccion"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Foto64    string `json:"foto64"`
	Latitud   string `json:"latitud"`
	Longitud  string `json:"longitud"`
}

type chatMessageInput struct {
	ConversationID string `json:"chatConversationID"`
	RecipientID    string `json:"chatRecipientID"`
	Text           string `json:"chatText"`
}

func (c *Controller) LoginRender(w http.ResponseWriter, r *http.Request) {
	if c.authenticated(r) {
		http.Redirect(w, r, ADMIN_INICIO, http.StatusSeeOther)
		return
	}

	ctx := templ.WithChildren(r.Context(), LoginRender())
	MainLayout("Acceso | Panel administrativo", Default(), xutility.AssetVersion).Render(ctx, w)
}

func (c *Controller) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	var signals loginSubmitInput
	if err := datastar.ReadSignals(r, &signals); err != nil {
		_ = r.ParseForm()
		signals.Username = r.FormValue("username")
		signals.Password = r.FormValue("password")
		signals.Context = r.FormValue("context")
	}

	result, err := xrepo.LoginV2(r.Context(), c.DB, model.NewLogin2{Username: signals.Username, Password: signals.Password})
	sse := datastar.NewSSE(w, r)
	if err != nil || result == nil {
		message := "No fue posible iniciar sesión."
		if err != nil {
			message = err.Error()
		}
		var body bytes.Buffer
		LoginError(message).Render(r.Context(), &body)
		selector := "#login-result"
		if r.URL.Query().Get("mode") == "relogin" {
			selector = "#relogin-result"
		}
		_ = sse.PatchElements(body.String(), datastar.WithSelector(selector), datastar.WithModeOuter())
		buttonID := "login-submit"
		if r.URL.Query().Get("mode") == "relogin" {
			buttonID = "relogin-submit"
		}
		_ = sse.ExecuteScript("const button = document.getElementById('" + buttonID + "'); if (button) { button.disabled = false; button.removeAttribute('aria-busy'); button.textContent = button.id === 'relogin-submit' ? 'Iniciar sesión' : 'Ingresar'; }")
		return
	}

	if r.URL.Query().Get("mode") == "relogin" {
		returnPath := r.URL.Query().Get("return_path")
		if !strings.HasPrefix(returnPath, "/adminx/") {
			returnPath = ADMIN_INICIO
		}
		contextValue := signals.Context
		if contextValue == "" {
			contextValue = r.URL.Query().Get("context")
		}
		if target, parseErr := url.Parse(returnPath); parseErr == nil {
			query := target.Query()
			if contextValue != "" {
				query.Set("context", contextValue)
			}
			target.RawQuery = query.Encode()
			_ = sse.Redirect(target.String())
			return
		}
		_ = sse.Redirect(ADMIN_INICIO)
		return
	}

	_ = sse.Redirect(ADMIN_INICIO)
}

func (c *Controller) Logout(w http.ResponseWriter, r *http.Request) {
	cookies := []string{
		"galletita_traviesa",
	}

	for _, name := range cookies {
		http.SetCookie(w, &http.Cookie{
			Name:   name,
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
	}

	http.Redirect(w, r, ADMIN_HOME, http.StatusSeeOther)
}

func (c *Controller) InicioRender(w http.ResponseWriter, r *http.Request) {
	ctx := templ.WithChildren(r.Context(), InicioRender(c.inicioData(r)))
	MainLayout("Acceso | Panel administrativo", Default(), xutility.AssetVersion).Render(ctx, w)
}

func (c *Controller) Chat(w http.ResponseWriter, r *http.Request) {
	auth, err := utilidad.CtxValue(r.Context(), c.DB, "")
	if err != nil || auth == nil || auth.SessionKey == nil || auth.SessionKey.UsuarioID == "" {
		http.Error(w, "Sesión no autorizada", http.StatusUnauthorized)
		return
	}

	if conversationID := strings.TrimSpace(r.URL.Query().Get("conversation_id")); conversationID != "" {
		c.chatConversation(w, r, auth.SessionKey.UsuarioID, conversationID)
		return
	}

	conversations, err := xrepo.ChatsByUser(r.Context(), c.DB, auth.SessionKey.UsuarioID)
	if err != nil {
		c.chatError(w, r, "No fue posible cargar tus conversaciones.")
		return
	}
	unread, _ := xrepo.ChatsNoLeidos(r.Context(), c.DB, auth.SessionKey.UsuarioID)
	var body bytes.Buffer
	if err := ChatConversationList(conversations).Render(r.Context(), &body); err != nil {
		c.chatError(w, r, "No fue posible renderizar tus conversaciones.")
		return
	}
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#chat-conversation-list"), datastar.WithModeOuter())
	patchChatUnreadBadges(sse, unread)
	_ = sse.ExecuteScript("document.getElementById('chat-conversation-list')?.classList.remove('hidden'); document.getElementById('chat-conversation-view')?.classList.add('hidden'); document.getElementById('chat-users-view')?.classList.add('hidden'); const title=document.getElementById('chat-dialog-title'); if(title) title.textContent='Chat';")
}

func (c *Controller) chatConversation(w http.ResponseWriter, r *http.Request, userID, conversationID string) {
	conversations, err := xrepo.ChatsByUser(r.Context(), c.DB, userID)
	if err != nil {
		c.chatError(w, r, "No fue posible cargar tus conversaciones.")
		return
	}
	var conversation *model.ResponseChatConversacion
	for _, item := range conversations {
		if item != nil && item.ID == conversationID {
			conversation = item
			break
		}
	}
	if conversation == nil {
		c.chatError(w, r, "La conversación no existe o no está disponible.")
		return
	}
	messages, err := xrepo.ChatMensajes(r.Context(), c.DB, conversationID)
	if err != nil {
		c.chatError(w, r, "No fue posible cargar los mensajes.")
		return
	}
	unread, _ := xrepo.ChatsNoLeidos(r.Context(), c.DB, userID)
	var body bytes.Buffer
	if err := ChatConversationView(conversation, messages, userID, chatConversationRecipient(conversation, userID), conversation.Nombre).Render(r.Context(), &body); err != nil {
		c.chatError(w, r, "No fue posible renderizar la conversación.")
		return
	}
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#chat-conversation-view"), datastar.WithModeOuter())
	patchChatUnreadBadges(sse, unread)
	_ = sse.MarshalAndPatchSignals(map[string]any{"chatConversationID": conversation.ID, "chatRecipientID": chatConversationRecipient(conversation, userID), "chatRecipientName": conversation.Nombre, "chatText": "", "chatSending": false})
	_ = sse.ExecuteScript("document.getElementById('chat-conversation-list')?.classList.add('hidden'); document.getElementById('chat-users-view')?.classList.add('hidden'); document.getElementById('chat-conversation-view')?.classList.remove('hidden'); const title=document.getElementById('chat-dialog-title'); if(title) title.textContent=" + strconv.Quote(conversation.Nombre) + ";")
}

func (c *Controller) ChatUsers(w http.ResponseWriter, r *http.Request) {
	if !c.authenticated(r) {
		http.Error(w, "Sesión no autorizada", http.StatusUnauthorized)
		return
	}
	auth, _ := utilidad.CtxValue(r.Context(), c.DB, "")
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
		if user == nil || user.ID == auth.SessionKey.UsuarioID {
			continue
		}
		filtered = append(filtered, user)
	}
	var body bytes.Buffer
	if err := ChatUserList(filtered).Render(r.Context(), &body); err != nil {
		c.chatError(w, r, "No fue posible renderizar los usuarios.")
		return
	}
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#chat-users"), datastar.WithModeOuter())
}

func (c *Controller) ChatCompose(w http.ResponseWriter, r *http.Request) {
	auth, err := utilidad.CtxValue(r.Context(), c.DB, "")
	if err != nil || auth == nil || auth.SessionKey == nil || auth.SessionKey.UsuarioID == "" {
		http.Error(w, "Sesión no autorizada", http.StatusUnauthorized)
		return
	}
	recipientID := strings.TrimSpace(r.URL.Query().Get("recipient_id"))
	recipientName := strings.TrimSpace(r.URL.Query().Get("name"))
	if recipientID == "" || recipientName == "" || recipientID == auth.SessionKey.UsuarioID {
		c.chatError(w, r, "El destinatario seleccionado no es válido.")
		return
	}
	var body bytes.Buffer
	_ = ChatConversationView(nil, nil, auth.SessionKey.UsuarioID, recipientID, recipientName).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#chat-conversation-view"), datastar.WithModeOuter())
	_ = sse.MarshalAndPatchSignals(map[string]any{"chatConversationID": "", "chatRecipientID": recipientID, "chatRecipientName": recipientName, "chatText": "", "chatSending": false})
	_ = sse.ExecuteScript("document.getElementById('chat-users-view')?.classList.add('hidden'); document.getElementById('chat-conversation-list')?.classList.add('hidden'); document.getElementById('chat-conversation-view')?.classList.remove('hidden'); const title=document.getElementById('chat-dialog-title'); if(title) title.textContent=" + strconv.Quote(recipientName) + ";")
}

func (c *Controller) ChatSend(w http.ResponseWriter, r *http.Request) {
	auth, err := utilidad.CtxValue(r.Context(), c.DB, "")
	if err != nil || auth == nil || auth.SessionKey == nil || auth.SessionKey.UsuarioID == "" {
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
		signals.RecipientID = strings.TrimSpace(r.URL.Query().Get("recipient_id"))
	}
	signals.Text = strings.TrimSpace(signals.Text)
	if signals.RecipientID == "" || signals.Text == "" {
		c.chatError(w, r, "Selecciona un destinatario y escribe un mensaje.")
		return
	}
	tipo := model.ChatMensajeTipoText
	message, err := xrepo.ChatEnviarMensaje(r.Context(), c.DB, model.ChatEnviarMensajeInput{SenderID: auth.SessionKey.UsuarioID, DestinatorID: signals.RecipientID, Tipo: tipo, Texto: signals.Text})
	if err != nil || message == nil {
		if err == nil {
			err = fmt.Errorf("no fue posible enviar el mensaje")
		}
		c.chatError(w, r, err.Error())
		return
	}
	conversationID := message.ConversacionID
	conversations, err := xrepo.ChatsByUser(r.Context(), c.DB, auth.SessionKey.UsuarioID)
	if err != nil {
		c.chatError(w, r, "El mensaje se envió, pero no se pudo actualizar la conversación.")
		return
	}
	var conversation *model.ResponseChatConversacion
	for _, item := range conversations {
		if item != nil && item.ID == conversationID {
			conversation = item
			break
		}
	}
	messages, err := xrepo.ChatMensajes(r.Context(), c.DB, conversationID)
	if err != nil || conversation == nil {
		c.chatError(w, r, "El mensaje se envió, pero no se pudo actualizar la conversación.")
		return
	}
	var body bytes.Buffer
	_ = ChatConversationView(conversation, messages, auth.SessionKey.UsuarioID, chatConversationRecipient(conversation, auth.SessionKey.UsuarioID), conversation.Nombre).Render(r.Context(), &body)
	unread, _ := xrepo.ChatsNoLeidos(r.Context(), c.DB, auth.SessionKey.UsuarioID)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#chat-conversation-view"), datastar.WithModeOuter())
	patchChatUnreadBadges(sse, unread)
	_ = sse.MarshalAndPatchSignals(map[string]any{"chatSending": false, "chatText": "", "chatConversationID": conversationID})
	_ = sse.ExecuteScript("const input=document.getElementById('chat-message-input'); if(input){input.value='';} document.getElementById('chat-conversation-list')?.classList.add('hidden'); document.getElementById('chat-users-view')?.classList.add('hidden'); document.getElementById('chat-conversation-view')?.classList.remove('hidden');")
}

func (c *Controller) chatError(w http.ResponseWriter, r *http.Request, message string) {
	sse := datastar.NewSSE(w, r)
	var body bytes.Buffer
	_ = ChatError(message).Render(r.Context(), &body)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#chat-error"), datastar.WithModeOuter())
	_ = sse.MarshalAndPatchSignals(map[string]bool{"chatSending": false})
}

func patchChatUnreadBadges(sse *datastar.ServerSentEventGenerator, count int32) {
	_ = sse.PatchElements(chatUnreadBadgeHTML("chat-unread-badge-trigger", count), datastar.WithSelector("#chat-unread-badge-trigger"), datastar.WithModeOuter())
	_ = sse.PatchElements(chatUnreadBadgeHTML("chat-unread-badge-menu", count), datastar.WithSelector("#chat-unread-badge-menu"), datastar.WithModeOuter())
}

func chatUnreadBadgeHTML(id string, count int32) string {
	if count <= 0 {
		return `<span id="` + id + `" data-chat-unread class="hidden rounded-full bg-primary px-1.5 py-0.5 text-[10px] font-semibold leading-none text-primary-foreground">0</span>`
	}
	return `<span id="` + id + `" data-chat-unread class="rounded-full bg-primary px-1.5 py-0.5 text-[10px] font-semibold leading-none text-primary-foreground">` + strconv.Itoa(int(count)) + `</span>`
}

func chatDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Local().Format("02/01/2006 15:04")
}

func chatUnreadString(count int32) string {
	if count > 99 {
		return "99+"
	}
	return strconv.Itoa(int(count))
}

func chatUnreadBadgeClass(count int32) string {
	if count <= 0 {
		return "hidden rounded-full bg-primary px-1.5 py-0.5 text-[10px] font-semibold leading-none text-primary-foreground"
	}
	return "rounded-full bg-primary px-1.5 py-0.5 text-[10px] font-semibold leading-none text-primary-foreground"
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

func chatConversationPhotoURL(conversation *model.ResponseChatConversacion) string {
	if conversation == nil || conversation.FotoURL == nil {
		return ""
	}
	return chatPathURL(*conversation.FotoURL)
}

func chatPathURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "/") {
		return value
	}
	return "/" + strings.TrimPrefix(value, "./")
}

func chatMessagesChronological(messages []*model.ChatMensaje) []*model.ChatMensaje {
	result := make([]*model.ChatMensaje, len(messages))
	for index, message := range messages {
		result[len(messages)-index-1] = message
	}
	return result
}

func chatMessageClass(message *model.ChatMensaje, currentUserID string) string {
	if message != nil && message.SenderID == currentUserID {
		return "ml-auto max-w-[85%] rounded-2xl rounded-br-sm bg-primary px-3 py-2 text-primary-foreground"
	}
	return "mr-auto max-w-[85%] rounded-2xl rounded-bl-sm bg-background px-3 py-2 text-foreground shadow-sm"
}

func chatConversationRecipient(conversation *model.ResponseChatConversacion, currentUserID string) string {
	if conversation == nil {
		return ""
	}
	for _, member := range conversation.Miembros {
		if member != nil && member.UsuarioID != currentUserID {
			return member.UsuarioID
		}
	}
	return ""
}

func chatConversationClick(id string) string {
	return "$chatConversationID=" + strconv.Quote(id) + ";$chatSending=false;$chatText='';@get('/adminx/chat?conversation_id='+encodeURIComponent($chatConversationID))"
}

func chatUserClick(user *model.Usuario) string {
	if user == nil {
		return ""
	}
	name := inicioUserName(&model.ResponseMe{Usuario: user})
	return "$chatRecipientID=" + strconv.Quote(user.ID) + ";$chatRecipientName=" + strconv.Quote(name) + ";$chatConversationID='';$chatText='';$chatSending=false;@get('/adminx/chat/compose?recipient_id='+encodeURIComponent($chatRecipientID)+'&name='+encodeURIComponent($chatRecipientName))"
}

func chatNewScript() string {
	return "$chatConversationID='';$chatRecipientID='';$chatRecipientName='';$chatText='';$chatSending=false;document.getElementById('chat-conversation-list')?.classList.add('hidden');document.getElementById('chat-conversation-view')?.classList.add('hidden');document.getElementById('chat-users-view')?.classList.remove('hidden');document.getElementById('chat-back-button')?.classList.remove('hidden');const title=document.getElementById('chat-dialog-title');if(title)title.textContent='Nuevo chat';@get('/adminx/chat/users')"
}

func chatBackScript() string {
	return "$chatConversationID='';$chatRecipientID='';$chatRecipientName='';$chatText='';$chatSending=false;document.getElementById('chat-users-view')?.classList.add('hidden');document.getElementById('chat-conversation-view')?.classList.add('hidden');document.getElementById('chat-conversation-list')?.classList.remove('hidden');document.getElementById('chat-back-button')?.classList.add('hidden');const title=document.getElementById('chat-dialog-title');if(title)title.textContent='Chat';@get('/adminx/chat')"
}

func chatNotificationHandler() string {
	return "const data=evt.detail;if(data?.tipo==='chat'){const id=data.datos?.conversation_id;if($chatConversationID&&id===$chatConversationID){@get('/adminx/chat?conversation_id='+encodeURIComponent($chatConversationID))}else{@get('/adminx/chat')}}"
}

func chatSearchRequestScript() string {
	return "const input=document.getElementById('chat-user-search');const value=input?.value||'';$chatSearch=value;@get('/adminx/chat/users?q='+encodeURIComponent(value))"
}

func chatTriggerAttributes() templ.Attributes {
	return templ.Attributes{
		"tabindex":                   "0",
		"type":                       "button",
		"data-slot":                  "dialog-trigger",
		"data-base-ui-click-trigger": true,
		"data-templ-controls":        "chat-dialog",
		"aria-haspopup":              "dialog",
		"aria-expanded":              "false",
		"data-on:click":              "$chatConversationID='';$chatText='';$chatSending=false;@get('/adminx/chat')",
	}
}

func (c *Controller) Perfil(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, ADMIN_INICIO, http.StatusSeeOther)
		return
	}

	auth, err := utilidad.CtxValue(r.Context(), c.DB, "update_perfil")
	if err != nil || auth == nil || auth.SessionKey == nil || auth.SessionKey.UsuarioID == "" {
		http.Error(w, "Sesión no autorizada", http.StatusUnauthorized)
		return
	}

	var signals perfilSubmitInput
	if err := datastar.ReadSignals(r, &signals); err != nil {
		_ = r.ParseForm()
		signals.Nombres = r.FormValue("nombres")
		signals.Apellido1 = r.FormValue("apellido1")
		signals.Apellido2 = r.FormValue("apellido2")
		signals.Documento = r.FormValue("documento")
		signals.Celular = r.FormValue("celular")
		signals.Correo = r.FormValue("correo")
		signals.Sexo = r.FormValue("sexo")
		signals.Direccion = r.FormValue("direccion")
		signals.Username = r.FormValue("username")
		signals.Password = r.FormValue("password")
		signals.Foto64 = r.FormValue("foto64")
		signals.Latitud = r.FormValue("latitud")
		signals.Longitud = r.FormValue("longitud")
	}

	input := model.UpdatePerfil{
		ID:        auth.SessionKey.UsuarioID,
		Nombres:   strings.TrimSpace(signals.Nombres),
		Apellido1: strings.TrimSpace(signals.Apellido1),
		Apellido2: perfilOptionalString(signals.Apellido2),
		Documento: perfilOptionalString(signals.Documento),
		Celular:   perfilOptionalString(signals.Celular),
		Correo:    perfilOptionalString(signals.Correo),
		Sexo:      perfilOptionalString(signals.Sexo),
		Direccion: perfilOptionalString(signals.Direccion),
		Username:  perfilOptionalString(signals.Username),
		Password:  perfilOptionalString(signals.Password),
		Foto64:    perfilOptionalString(signals.Foto64),
	}
	if input.Nombres == "" || input.Apellido1 == "" {
		c.perfilError(w, r, "Nombres y primer apellido son obligatorios.")
		return
	}
	if signals.Latitud != "" {
		value, parseErr := strconv.ParseFloat(strings.TrimSpace(signals.Latitud), 64)
		if parseErr != nil {
			c.perfilError(w, r, "La latitud no es válida.")
			return
		}
		input.Latitud = &value
	}
	if signals.Longitud != "" {
		value, parseErr := strconv.ParseFloat(strings.TrimSpace(signals.Longitud), 64)
		if parseErr != nil {
			c.perfilError(w, r, "La longitud no es válida.")
			return
		}
		input.Longitud = &value
	}

	if _, err := xrepo.UpdatePerfil(r.Context(), c.DB, input); err != nil {
		c.perfilError(w, r, err.Error())
		return
	}

	sse := datastar.NewSSE(w, r)
	_ = sse.MarshalAndPatchSignals(map[string]bool{"perfilSaving": false, "perfilSaved": true})
}

func (c *Controller) perfilError(w http.ResponseWriter, r *http.Request, message string) {
	var body bytes.Buffer
	_ = PerfilFormError(message).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#perfil-form-error"), datastar.WithModeOuter())
	_ = sse.MarshalAndPatchSignals(map[string]bool{"perfilSaving": false, "perfilSaved": false})
}

func perfilOptionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func (c *Controller) ConnectedUsers(w http.ResponseWriter, r *http.Request) {
	if !c.authenticated(r) {
		http.Error(w, "Sesión no autorizada", http.StatusUnauthorized)
		return
	}

	users, err := xrepo.UsuariosConectados(r.Context(), c.DB)
	message := ""
	if err != nil {
		message = "No hay usuarios conectados."
	}

	var body bytes.Buffer
	if renderErr := xutility.ConnectedUsersContent(users, message).Render(r.Context(), &body); renderErr != nil {
		http.Error(w, "No fue posible cargar los usuarios conectados", http.StatusInternalServerError)
		return
	}

	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#connected-users-content"), datastar.WithModeOuter())
}

func (c *Controller) ActiveAlerts(w http.ResponseWriter, r *http.Request) {
	if !c.authenticated(r) {
		http.Error(w, "Sesión no autorizada", http.StatusUnauthorized)
		return
	}

	alerts, err := xrepo.Notificaciones(r.Context(), c.DB)
	if err != nil {
		http.Error(w, "No fue posible cargar los avisos activos", http.StatusInternalServerError)
		return
	}

	var body bytes.Buffer
	if renderErr := xutility.ActiveAlertsContent(alerts).Render(r.Context(), &body); renderErr != nil {
		http.Error(w, "No fue posible renderizar los avisos activos", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(body.Bytes())
}

func (c *Controller) renderModuleRoute(w http.ResponseWriter, r *http.Request, module http.Handler) {
	if r.Header.Get("X-Requested-With") == "fetch" || r.Header.Get("Datastar-Request") == "true" {
		module.ServeHTTP(w, r)
		return
	}

	recorder := httptest.NewRecorder()
	module.ServeHTTP(recorder, r)
	data := c.inicioData(r)
	content := append([]byte(nil), recorder.Body.Bytes()...)
	data.Content = templ.ComponentFunc(func(_ context.Context, target io.Writer) error {
		_, err := target.Write(content)
		return err
	})

	if recorder.Code >= http.StatusBadRequest {
		w.WriteHeader(recorder.Code)
	}
	ctx := templ.WithChildren(r.Context(), InicioRender(data))
	_ = MainLayout("Panel administrativo", Default(), xutility.AssetVersion).Render(ctx, w)
}

func (c *Controller) ContextSubmit(w http.ResponseWriter, r *http.Request) {
	var input contextInput
	if err := datastar.ReadSignals(r, &input); err != nil {
		_ = r.ParseForm()
		input.Context = r.FormValue("context")
	}

	parts := strings.SplitN(input.Context, "|", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		c.contextError(w, r, "La unidad y el rol seleccionados no son válidos.")
		return
	}

	unidadID, rolID := parts[0], parts[1]
	me, err := xrepo.Me(r.Context(), c.DB, model.InputMe{UnidadID: unidadID})
	if err != nil || me == nil || !inicioHasRole(me, unidadID, rolID) {
		c.contextError(w, r, "No tienes acceso a la unidad y rol seleccionados.")
		return
	}

	menus, err := xrepo.GetMenusbyRol(r.Context(), c.DB, rolID)
	if err != nil {
		c.contextError(w, r, "No fue posible cargar los menús del rol seleccionado.")
		return
	}

	cookie, err := utilidad.CtxGetCookie(r)
	if err != nil {
		c.contextError(w, r, "La sesión ya no está disponible. Inicia sesión nuevamente.")
		return
	}
	cookie.RolID = rolID
	cookie.UnidadID = unidadID
	utilidad.CtxSetCookie(r.Context(), *cookie)

	data := InicioData{Authenticated: true, Me: me, ChatUnread: chatUnread(r.Context(), c.DB, cookie.UserID), UnidadID: unidadID, RolID: rolID, ActivePath: inicioActivePath(r)}
	data.Me.Menus = menus

	var body bytes.Buffer
	if err := InicioSidebar(data).Render(r.Context(), &body); err != nil {
		c.contextError(w, r, "No fue posible actualizar el menú lateral.")
		return
	}

	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#admin-sidebar"), datastar.WithModeOuter())
	_ = sse.ExecuteScript("const role = document.getElementById('active-role'); if (role) role.textContent = " + strconv.Quote(inicioSelectedRole(me, unidadID, rolID)) + "; const result = document.getElementById('context-result'); if (result) result.replaceChildren(); const selector = document.querySelector('[data-context-selector]'); if (selector) { selector.removeAttribute('aria-busy'); selector.querySelectorAll('input,button').forEach(element => element.disabled = false); }")
}

func chatUnread(ctx context.Context, db *sql.DB, userID string) int32 {
	if userID == "" {
		return 0
	}
	count, err := xrepo.ChatsNoLeidos(ctx, db, userID)
	if err != nil || count < 0 {
		return 0
	}
	return count
}

func (c *Controller) contextError(w http.ResponseWriter, r *http.Request, message string) {
	var body bytes.Buffer
	_ = InicioContextError(message).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#context-result"), datastar.WithModeOuter())
	_ = sse.ExecuteScript("const selector = document.querySelector('[data-context-selector]'); if (selector) { selector.removeAttribute('aria-busy'); selector.querySelectorAll('input,button').forEach(element => element.disabled = false); }")
}

func inicioHasRole(me *model.ResponseMe, unidadID, rolID string) bool {
	if me == nil {
		return false
	}
	for _, role := range me.Roles {
		if role != nil && role.Unidad != nil && role.Rol != nil && role.Unidad.ID == unidadID && role.Rol.ID == rolID {
			return true
		}
	}
	return false
}

func (c *Controller) inicioData(r *http.Request) InicioData {
	auth, err := utilidad.CtxValue(r.Context(), c.DB, "")
	if err != nil {
		return InicioData{}
	}

	data := InicioData{
		Authenticated: true,
		ChatUnread:    chatUnread(r.Context(), c.DB, auth.SessionKey.UsuarioID),
		UnidadID:      auth.UnidadID,
		RolID:         auth.RolID,
		ActivePath:    inicioActivePath(r),
	}
	data.Me, _ = xrepo.Me(r.Context(), c.DB, model.InputMe{UnidadID: auth.UnidadID})
	requestedContext := r.URL.Query().Get("context")
	if requestedContext != "" {
		parts := strings.SplitN(requestedContext, "|", 2)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			requestedMe, requestedErr := xrepo.Me(r.Context(), c.DB, model.InputMe{UnidadID: parts[0]})
			if requestedErr == nil && inicioHasRole(requestedMe, parts[0], parts[1]) {
				data.Me = requestedMe
				data.UnidadID = parts[0]
				data.RolID = parts[1]
			}
		}
	}

	if data.Me != nil && !inicioHasRole(data.Me, data.UnidadID, data.RolID) {
		if firstUnit, firstRole, ok := inicioFirstRole(data.Me); ok {
			data.UnidadID = firstUnit
			data.RolID = firstRole
		}
	}
	if requestedContext != "" || data.UnidadID != auth.UnidadID || data.RolID != auth.RolID {
		if cookie, cookieErr := utilidad.CtxGetCookie(r); cookieErr == nil {
			cookie.UnidadID = data.UnidadID
			cookie.RolID = data.RolID
			utilidad.CtxSetCookie(r.Context(), *cookie)
		}
	}
	if data.Me != nil {
		// La cookie define el rol activo. Reemplazamos los menús generales
		// devueltos por Me por los que pertenecen a ese rol concreto.
		if menus, menuErr := xrepo.GetMenusbyRol(r.Context(), c.DB, data.RolID); menuErr == nil {
			data.Me.Menus = menus
		}
	}
	return data
}

func inicioFirstRole(me *model.ResponseMe) (string, string, bool) {
	if me == nil {
		return "", "", false
	}
	for _, role := range me.Roles {
		if role != nil && role.Unidad != nil && role.Rol != nil {
			return role.Unidad.ID, role.Rol.ID, true
		}
	}
	return "", "", false
}

func inicioContextSignal(unidadID, rolID string) string {
	return strconv.Quote(unidadID + "|" + rolID)
}

func (c *Controller) authenticated(r *http.Request) bool {
	_, err := utilidad.CtxValue(r.Context(), c.DB, "")
	return err == nil
}

func inicioUserName(me *model.ResponseMe) string {
	if me == nil || me.Usuario == nil {
		return "Usuario"
	}
	return strings.TrimSpace(me.Usuario.Nombres + " " + me.Usuario.Apellido1)
}

func inicioUser(me *model.ResponseMe) *model.Usuario {
	if me == nil {
		return nil
	}
	return me.Usuario
}

func inicioUserInitials(me *model.ResponseMe) string {
	name := inicioUserName(me)
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "U"
	}
	if len(parts) == 1 {
		return strings.ToUpper(string([]rune(parts[0])[0]))
	}
	return strings.ToUpper(string([]rune(parts[0])[0]) + string([]rune(parts[len(parts)-1])[0]))
}

func perfilUserString(user *model.Usuario, field string) string {
	if user == nil {
		return ""
	}
	switch field {
	case "nombres":
		return user.Nombres
	case "apellido1":
		return user.Apellido1
	case "apellido2":
		return perfilStringValue(user.Apellido2)
	case "documento":
		return perfilStringValue(user.Documento)
	case "celular":
		return perfilStringValue(user.Celular)
	case "correo":
		return perfilStringValue(user.Correo)
	case "sexo":
		return perfilStringValue(user.Sexo)
	case "direccion":
		return perfilStringValue(user.Direccion)
	case "username":
		return user.Username
	case "latitud":
		if user.Latitud != nil {
			return strconv.FormatFloat(*user.Latitud, 'f', -1, 64)
		}
	case "longitud":
		if user.Longitud != nil {
			return strconv.FormatFloat(*user.Longitud, 'f', -1, 64)
		}
	}
	return ""
}

func perfilStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func perfilSignalString(value string) string {
	return strconv.Quote(value)
}

func perfilBind(signal string) templ.Attributes {
	return templ.Attributes{"data-bind:" + signal: true}
}

func perfilPhotoURL(user *model.Usuario) string {
	if user == nil || user.FotoURL == nil || strings.TrimSpace(*user.FotoURL) == "" {
		return ""
	}
	path := strings.TrimSpace(*user.FotoURL)
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "/") {
		return path
	}
	return "/" + strings.TrimPrefix(path, "./")
}

func perfilTriggerAttributes(user *model.Usuario) templ.Attributes {
	return templ.Attributes{
		"tabindex":                   "0",
		"type":                       "button",
		"data-slot":                  "dialog-trigger",
		"data-base-ui-click-trigger": true,
		"data-templ-controls":        "perfil-dialog",
		"aria-haspopup":              "dialog",
		"aria-expanded":              "false",
		"data-on:click":              perfilResetScript(user),
	}
}

func perfilResetScript(user *model.Usuario) string {
	assign := func(signal, field string) string {
		return "$" + signal + "=" + strconv.Quote(perfilUserString(user, field))
	}
	return strings.Join([]string{
		assign("nombres", "nombres"), assign("apellido1", "apellido1"), assign("apellido2", "apellido2"),
		assign("documento", "documento"), assign("celular", "celular"), assign("correo", "correo"),
		assign("sexo", "sexo"), assign("direccion", "direccion"), assign("username", "username"),
		assign("latitud", "latitud"), assign("longitud", "longitud"),
		"$password='';$foto64='';$perfilSaving=false;$perfilSaved=false;const error=document.getElementById('perfil-form-error');if(error){error.replaceChildren();}const file=document.getElementById('perfil-foto');if(file){file.value='';}const image=document.getElementById('perfil-photo-preview');if(image){image.src=" + strconv.Quote(perfilPhotoURL(user)) + ";image.hidden=!image.src;}",
	}, ";")
}

func inicioSelectedRole(me *model.ResponseMe, unidadID, rolID string) string {
	if me != nil {
		for _, role := range me.Roles {
			if role.Unidad != nil && role.Rol != nil && role.Unidad.ID == unidadID && role.Rol.ID == rolID {
				return role.Unidad.Nombre + " / " + role.Rol.Nombre
			}
		}
	}
	return "Seleccionar unidad / rol"
}

func inicioRoleItems(me *model.ResponseMe) []combobox.ItemData {
	if me == nil {
		return nil
	}

	items := make([]combobox.ItemData, 0, len(me.Roles))
	for _, role := range me.Roles {
		if role.Unidad == nil || role.Rol == nil {
			continue
		}
		items = append(items, combobox.ItemData{
			Value: role.Unidad.ID + "|" + role.Rol.ID,
			Label: role.Unidad.Nombre + " / " + role.Rol.Nombre,
		})
	}
	return items
}

func inicioRootMenus(menus []*model.Menus) []*model.Menus {
	roots := make([]*model.Menus, 0)
	for _, menu := range menus {
		if menu != nil && menu.PadreID == nil {
			roots = append(roots, menu)
		}
	}
	return roots
}

func inicioChildMenus(menus []*model.Menus, parentID string) []*model.Menus {
	children := make([]*model.Menus, 0)
	for _, menu := range menus {
		if menu != nil && menu.PadreID != nil && *menu.PadreID == parentID {
			children = append(children, menu)
		}
	}
	return children
}

func inicioActivePath(r *http.Request) string {
	path := r.URL.Path
	if path != ADMIN_CONTEXT {
		return path
	}
	if referer, err := url.Parse(r.Referer()); err == nil && strings.HasPrefix(referer.Path, "/adminx/") {
		return referer.Path
	}
	return ADMIN_INICIO
}

func inicioMenuClass(path, activePath, base string) string {
	if path == activePath {
		return base + " bg-sidebar-foreground/15 text-sidebar-foreground"
	}
	return base
}

func inicioMenuExpanded(children []*model.Menus, activePath string) string {
	for _, child := range children {
		if child != nil && child.Path == activePath {
			return "space-y-1 border-l border-sidebar-border pl-3"
		}
	}
	return "hidden space-y-1 border-l border-sidebar-border pl-3"
}

func inicioMenuClick(menu *model.Menus) string {
	if menu == nil {
		return ""
	}

	path := strconv.Quote(menu.Path)
	script := `(async()=>{const selected=event.currentTarget;const menuItems=document.querySelectorAll('[data-menu-path]');menuItems.forEach(item=>item.disabled=true);const target=document.getElementById('admin-content');if(!target){menuItems.forEach(item=>item.disabled=false);return}const path=__PATH__;target.setAttribute('aria-busy','true');target.innerHTML='<div class="flex min-h-48 items-center justify-center gap-2 text-sm text-muted-foreground"><span class="size-4 animate-spin rounded-full border-2 border-current border-t-transparent"></span>Cargando, espere...</div>';try{const response=await fetch(path,{headers:{'Accept':'text/html','X-Requested-With':'fetch'}});const html=await response.text();const parsed=new DOMParser().parseFromString(html,'text/html');const page=parsed.querySelector('#admin-content');target.innerHTML=page?page.innerHTML:html;if(response.ok){window.history.pushState({},'',path);document.querySelectorAll('[data-menu-path]').forEach(item=>item.classList.remove('bg-sidebar-foreground/15','text-sidebar-foreground'));selected.classList.add('bg-sidebar-foreground/15','text-sidebar-foreground');selected.closest('[id^="admin-menu-"]')?.classList.remove('hidden');if(window.matchMedia('(max-width: 767px)').matches){const shell=document.getElementById('admin-shell');shell?.setAttribute('data-sidebar-collapsed','true');document.getElementById('admin-sidebar')?.classList.add('hidden');document.getElementById('admin-sidebar')?.classList.remove('flex')}}}catch(error){target.innerHTML='<div class="rounded-lg border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">No fue posible cargar el contenido: '+error.message+'</div>'}finally{target.removeAttribute('aria-busy');menuItems.forEach(item=>item.disabled=false)}})();`
	return strings.Replace(script, "__PATH__", path, 1)
}
