package tickets

import (
	"encoding/json"
	"html"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/starfederation/datastar-go/datastar"
)

func ticketModalRequestScript(path string) string {
	return "const button=evt.currentTarget;if(button){button.disabled=true;button.setAttribute('aria-busy','true');}@get('" + path + "')"
}

func ticketModalRequestWithIDScript(path, id string) string {
	return "const button=evt.currentTarget;if(button){button.disabled=true;button.setAttribute('aria-busy','true');}@get('" + path + "?ticketid=" + id + "')"
}

func ticketModalLoadingScript() string {
	return "const content=document.getElementById('ticket-dialog-content');if(content){content.className='flex min-h-40 items-center justify-center gap-2 text-sm text-muted-foreground';content.innerHTML='<span class=\"size-4 animate-spin rounded-full border-2 border-current border-t-transparent\"></span>Cargando, espere...';}document.querySelectorAll('[data-ticket-modal-action]').forEach((button)=>{button.disabled=false;button.removeAttribute('aria-busy');});"
}

func ticketModalReadyScript() string {
	return "document.querySelectorAll('[data-ticket-modal-action]').forEach((button)=>{button.disabled=false;button.removeAttribute('aria-busy');});"
}

func ticketModalTitleScript(title string) string {
	data, _ := json.Marshal(title)
	return "document.querySelector('#ticket-dialog [data-ticket-title]')?.replaceChildren(document.createTextNode(" + string(data) + "));"
}

func ticketModalDescriptionScript(description string) string {
	data, _ := json.Marshal(description)
	return "document.querySelector('#ticket-dialog [data-templ-dialog-description]')?.replaceChildren(document.createTextNode(" + string(data) + "));"
}

func ticketRequestID(r *http.Request) string {
	id := strings.TrimSpace(r.URL.Query().Get("ticketid"))
	if id == "" {
		id = strings.TrimSpace(r.URL.Query().Get("id"))
	}
	return id
}

func ticketFormSubmitScript() string {
	return "$ticketSaving=true; @post('" + ADMIN_TICKETS_CREAR + "?estado='+encodeURIComponent($ticketEstado||''))"
}

func ticketReplySubmitScript(id string) string {
	return "$ticketSaving=true; @post('" + ADMIN_TICKETS_RESPONDER + "?ticketid=" + id + "&estado='+encodeURIComponent($ticketEstado||''))"
}

func ticketCloseScript(id string) string {
	message, _ := json.Marshal("¿Deseas cerrar este ticket? Esta acción no se puede deshacer.")
	return "if(window.confirm(" + string(message) + ")){$ticketSaving=true;@post('" + ADMIN_TICKETS_CERRAR + "?ticketid=" + id + "&estado='+encodeURIComponent($ticketEstado||''))}"
}

func ticketFormErrorWithID(id, message string) string {
	return "<div id=\"" + html.EscapeString(id) + "\" role=\"alert\" class=\"rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive\">" + html.EscapeString(message) + "</div>"
}

func (c *Controller) TicketFormularioRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	c.patchTicketModal(w, r, TicketCreateForm(), ticketModalTitleScript("Nuevo ticket")+ticketModalDescriptionScript("Describe el problema o solicitud que deseas registrar.")+ticketModalReadyScript())
}

func (c *Controller) TicketVisualizarRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	id := ticketRequestID(r)
	if id == "" {
		c.patchTicketModal(w, r, TicketModalError("No se recibió el identificador del ticket."), ticketModalTitleScript("Ticket")+ticketModalDescriptionScript("Falta el identificador para consultar la información.")+ticketModalReadyScript())
		return
	}
	detail, err := xrepo.VerTicketConRespuestas(r.Context(), c.DB, id)
	if err != nil || detail == nil {
		c.patchTicketModal(w, r, TicketModalError("No fue posible cargar el ticket."), ticketModalTitleScript("Ticket")+ticketModalDescriptionScript("No fue posible cargar la información.")+ticketModalReadyScript())
		return
	}
	c.patchTicketModal(w, r, TicketView(detail), ticketModalTitleScript("Visualizar ticket")+ticketModalDescriptionScript("Consulta el problema y el historial de respuestas.")+ticketModalReadyScript())
}

func (c *Controller) patchTicketModal(w http.ResponseWriter, r *http.Request, component templ.Component, after string) {
	var body strings.Builder
	if err := component.Render(r.Context(), &body); err != nil {
		http.Error(w, "No fue posible renderizar el modal", http.StatusInternalServerError)
		return
	}
	if r.Header.Get("Datastar-Request") != "true" {
		_, _ = w.Write([]byte(body.String()))
		return
	}
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#ticket-dialog-content"), datastar.WithModeOuter())
	if after != "" {
		_ = sse.ExecuteScript(after)
	}
}
