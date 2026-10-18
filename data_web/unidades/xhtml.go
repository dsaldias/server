package unidades

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

func unidadModalRequestScript(path string) string {
	return "const button=evt.currentTarget;if(button){button.disabled=true;button.setAttribute('aria-busy','true');} @get('" + path + "')"
}

func unidadModalRequestWithIDScript(path string) string {
	return "const button=evt.currentTarget;const unidadID=button?.getAttribute('data-unidad-id')||'';if(button){button.disabled=true;button.setAttribute('aria-busy','true');} @get('" + path + "?unidadid='+encodeURIComponent(unidadID))"
}

func unidadFormSubmitScript(unidad *model.Unidad) string {
	path := ADMIN_UNIDADES_CREAR
	if unidad != nil {
		path = ADMIN_UNIDADES_EDITAR
	}
	return "$unidadSaving=true; @post('" + path + "?pagina='+$pagina+'&size='+$size+'&nombre='+encodeURIComponent($nombre||''))"
}

func unidadModalLoading() string {
	return "const content=document.getElementById('unidad-dialog-content');if(content){content.className='flex min-h-40 items-center justify-center gap-2 text-sm text-muted-foreground';content.innerHTML='<span class=\"size-4 animate-spin rounded-full border-2 border-current border-t-transparent\"></span>Cargando, espere...';}document.querySelectorAll('[data-unidad-modal-action]').forEach((button)=>{button.disabled=false;button.removeAttribute('aria-busy');});"
}

func unidadModalReadyScript() string {
	return "document.querySelectorAll('[data-unidad-modal-action]').forEach((button)=>{button.disabled=false;button.removeAttribute('aria-busy');});"
}

func unidadModalTitleScript(title string) string {
	data, _ := json.Marshal(title)
	return "document.querySelector('#unidad-dialog [data-unidad-title]')?.replaceChildren(document.createTextNode(" + string(data) + "));"
}

func unidadModalDescriptionScript(description string) string {
	data, _ := json.Marshal(description)
	return "document.querySelector('#unidad-dialog [data-templ-dialog-description]')?.replaceChildren(document.createTextNode(" + string(data) + "));"
}

func unidadRequestID(r *http.Request) string {
	id := strings.TrimSpace(r.URL.Query().Get("unidadid"))
	if id == "" {
		id = strings.TrimSpace(r.URL.Query().Get("id"))
	}
	return id
}

func unidadFormMode(unidad *model.Unidad) string {
	if unidad == nil {
		return "create"
	}
	return "edit"
}

func unidadFormID(unidad *model.Unidad) string {
	if unidad == nil {
		return ""
	}
	return unidad.ID
}

func unidadFormValue(unidad *model.Unidad, field string) string {
	if unidad == nil {
		return ""
	}
	switch field {
	case "nombre":
		return unidad.Nombre
	case "descripcion":
		return optionalStringValue(unidad.Descripcion)
	case "latitud":
		return optionalFloatValue(unidad.Latitud)
	case "longitud":
		return optionalFloatValue(unidad.Longitud)
	default:
		return ""
	}
}

func unidadFormOrden(unidad *model.Unidad) string {
	if unidad == nil {
		return "0"
	}
	return unidadSignalInt(unidad.Orden)
}

func (c *Controller) UnidadFormularioRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	c.patchUnidadModal(w, r, UnidadForm(nil), unidadModalTitleScript("Nueva unidad")+unidadModalDescriptionScript("Registra la información y ubicación opcional de la unidad.")+unidadModalReadyScript())
}

func (c *Controller) UnidadVisualizarRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	id := unidadRequestID(r)
	if id == "" {
		c.patchUnidadModal(w, r, UnidadModalError("No se recibió el identificador de la unidad."), unidadModalTitleScript("Unidad")+unidadModalDescriptionScript("Falta el identificador para consultar la información.")+unidadModalReadyScript())
		return
	}
	unida, err := xrepo.UnidadByID(r.Context(), c.DB, id)
	if err != nil || unida == nil {
		c.patchUnidadModal(w, r, UnidadModalError("No fue posible cargar la unidad."), unidadModalTitleScript("Unidad")+unidadModalDescriptionScript("No fue posible cargar la información.")+unidadModalReadyScript())
		return
	}
	c.patchUnidadModal(w, r, UnidadView(unida), unidadModalTitleScript("Visualizar unidad")+unidadModalDescriptionScript("Consulta la información registrada de la unidad.")+unidadModalReadyScript())
}

func (c *Controller) UnidadEditarRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	id := unidadRequestID(r)
	if id == "" {
		c.patchUnidadModal(w, r, UnidadModalError("No se recibió el identificador de la unidad."), unidadModalTitleScript("Editar unidad")+unidadModalDescriptionScript("Falta el identificador para cargar la información.")+unidadModalReadyScript())
		return
	}
	unida, err := xrepo.UnidadByID(r.Context(), c.DB, id)
	if err != nil || unida == nil {
		c.patchUnidadModal(w, r, UnidadModalError("No fue posible cargar la unidad."), unidadModalTitleScript("Editar unidad")+unidadModalDescriptionScript("No fue posible cargar la información.")+unidadModalReadyScript())
		return
	}
	c.patchUnidadModal(w, r, UnidadForm(unida), unidadModalTitleScript("Editar unidad")+unidadModalDescriptionScript("Actualiza la información y ubicación opcional de la unidad.")+unidadModalReadyScript())
}

func (c *Controller) patchUnidadModal(w http.ResponseWriter, r *http.Request, component templ.Component, after string) {
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
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#unidad-dialog-content"), datastar.WithModeOuter())
	if after != "" {
		_ = sse.ExecuteScript(after)
	}
}
