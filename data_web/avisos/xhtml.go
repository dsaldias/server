package avisos

import (
	"encoding/json"
	"html"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

func avisoModalRequestScript(path string) string {
	return "const button=evt.currentTarget;if(button){button.disabled=true;button.setAttribute('aria-busy','true');}@get('" + path + "')"
}

func avisoModalRequestWithIDScript(path, id string) string {
	return "const button=evt.currentTarget;if(button){button.disabled=true;button.setAttribute('aria-busy','true');}@get('" + path + "?avisoid=" + id + "')"
}

func avisoModalLoadingScript() string {
	return "const content=document.getElementById('aviso-dialog-content');if(content){content.className='flex min-h-40 items-center justify-center gap-2 text-sm text-muted-foreground';content.innerHTML='<span class=\"size-4 animate-spin rounded-full border-2 border-current border-t-transparent\"></span>Cargando, espere...';}document.querySelectorAll('[data-aviso-modal-action]').forEach((button)=>{button.disabled=false;button.removeAttribute('aria-busy');});"
}

func avisoModalReadyScript() string {
	return "document.querySelectorAll('[data-aviso-modal-action]').forEach((button)=>{button.disabled=false;button.removeAttribute('aria-busy');});"
}

func avisoModalTitleScript(title string) string {
	data, _ := json.Marshal(title)
	return "document.querySelector('#aviso-dialog [data-aviso-title]')?.replaceChildren(document.createTextNode(" + string(data) + "));"
}

func avisoModalDescriptionScript(description string) string {
	data, _ := json.Marshal(description)
	return "document.querySelector('#aviso-dialog [data-templ-dialog-description]')?.replaceChildren(document.createTextNode(" + string(data) + "));"
}

func avisoRequestID(r *http.Request) string {
	id := strings.TrimSpace(r.URL.Query().Get("avisoid"))
	if id == "" {
		id = strings.TrimSpace(r.URL.Query().Get("id"))
	}
	return id
}

func avisoFormSubmitScript(editing bool) string {
	path := ADMIN_AVISOS_CREAR
	if editing {
		path = ADMIN_AVISOS_EDITAR
	}
	return "$avisoTimezone=Intl.DateTimeFormat().resolvedOptions().timeZone;$avisoSaving=true;@post('" + path + "')"
}

func avisoFormValue(aviso *model.Notificacion, field string) string {
	if aviso == nil {
		return ""
	}
	switch field {
	case "mensaje":
		return aviso.Mensaje
	case "desde":
		return avisoDateISO(aviso.Desde)
	case "hasta":
		return avisoDateISO(aviso.Hasta)
	default:
		return ""
	}
}

func avisoFormID(aviso *model.Notificacion) string {
	if aviso == nil {
		return ""
	}
	return aviso.ID
}

func avisoFormMode(aviso *model.Notificacion) string {
	if aviso == nil {
		return "create"
	}
	return "edit"
}

func avisoFormError(message string) string {
	return "<div id=\"aviso-form-error\" role=\"alert\" class=\"rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive\">" + html.EscapeString(message) + "</div>"
}

func (c *Controller) AvisoFormularioRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	c.patchAvisoModal(w, r, AvisoForm(nil), avisoModalTitleScript("Nuevo aviso")+avisoModalDescriptionScript("Define el mensaje y el periodo en el que estará activo.")+avisoModalReadyScript())
}

func (c *Controller) AvisoVisualizarRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	id := avisoRequestID(r)
	if id == "" {
		c.patchAvisoModal(w, r, AvisoModalError("No se recibió el identificador del aviso."), avisoModalTitleScript("Aviso")+avisoModalDescriptionScript("Falta el identificador para consultar la información.")+avisoModalReadyScript())
		return
	}
	aviso, err := xrepo.Notificacion(r.Context(), c.DB, id)
	if err != nil || aviso == nil {
		c.patchAvisoModal(w, r, AvisoModalError("No fue posible cargar el aviso."), avisoModalTitleScript("Aviso")+avisoModalDescriptionScript("No fue posible cargar la información.")+avisoModalReadyScript())
		return
	}
	c.patchAvisoModal(w, r, AvisoView(aviso), avisoModalTitleScript("Visualizar aviso")+avisoModalDescriptionScript("Consulta la información registrada del aviso.")+avisoModalReadyScript())
}

func (c *Controller) AvisoEditarRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	id := avisoRequestID(r)
	if id == "" {
		c.patchAvisoModal(w, r, AvisoModalError("No se recibió el identificador del aviso."), avisoModalTitleScript("Editar aviso")+avisoModalDescriptionScript("Falta el identificador para cargar la información.")+avisoModalReadyScript())
		return
	}
	aviso, err := xrepo.Notificacion(r.Context(), c.DB, id)
	if err != nil || aviso == nil {
		c.patchAvisoModal(w, r, AvisoModalError("No fue posible cargar el aviso."), avisoModalTitleScript("Editar aviso")+avisoModalDescriptionScript("No fue posible cargar la información.")+avisoModalReadyScript())
		return
	}
	c.patchAvisoModal(w, r, AvisoForm(aviso), avisoModalTitleScript("Editar aviso")+avisoModalDescriptionScript("Actualiza el mensaje y el periodo del aviso.")+avisoModalReadyScript())
}

func (c *Controller) patchAvisoModal(w http.ResponseWriter, r *http.Request, component templ.Component, after string) {
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
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#aviso-dialog-content"), datastar.WithModeOuter())
	if after != "" {
		_ = sse.ExecuteScript(after)
	}
}
