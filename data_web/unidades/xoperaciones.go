package unidades

import (
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type unidadFormSignals struct {
	UnidadNombre string `json:"unidadNombre"`
	Descripcion  string `json:"descripcion"`
	Orden        int32  `json:"orden"`
	Latitud      string `json:"latitud"`
	Longitud     string `json:"longitud"`
	ID           string `json:"unidadID"`
}

func (c *Controller) UnidadCrear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	c.unidadSave(w, r, false)
}

func (c *Controller) UnidadEditar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	c.unidadSave(w, r, true)
}

func (c *Controller) unidadSave(w http.ResponseWriter, r *http.Request, editing bool) {
	var signals unidadFormSignals
	if err := datastar.ReadSignals(r, &signals); err != nil {
		_ = r.ParseForm()
		signals.UnidadNombre = r.FormValue("nombre")
		signals.Descripcion = r.FormValue("descripcion")
		signals.Orden = positiveInt(r.FormValue("orden"), 0)
		signals.Latitud = r.FormValue("latitud")
		signals.Longitud = r.FormValue("longitud")
		signals.ID = r.FormValue("unidadID")
	}
	input := model.NewUnidad{
		Nombre:      strings.TrimSpace(signals.UnidadNombre),
		Descripcion: optionalString(signals.Descripcion),
		Orden:       signals.Orden,
		Latitud:     optionalFloat(signals.Latitud),
		Longitud:    optionalFloat(signals.Longitud),
	}
	var err error
	if editing {
		if strings.TrimSpace(signals.ID) == "" {
			err = fmt.Errorf("el identificador de la unidad es obligatorio para editar")
		} else {
			_, err = xrepo.UpdateUnidad(r.Context(), c.DB, model.UpdUnidad{ID: signals.ID, Nombre: input.Nombre, Descripcion: input.Descripcion, Orden: input.Orden, Latitud: input.Latitud, Longitud: input.Longitud})
		}
	} else {
		_, err = xrepo.CreateUnidad(r.Context(), c.DB, input)
	}
	sse := datastar.NewSSE(w, r)
	if err != nil {
		_ = sse.PatchElements(unidadFormError(err.Error()), datastar.WithSelector("#unidad-form-error"), datastar.WithModeOuter())
		_ = sse.MarshalAndPatchSignals(map[string]bool{"unidadSaving": false, "unidadSaved": false, "unidadesLoading": false})
		return
	}
	data := c.unidadesTableData(r)
	var body strings.Builder
	_ = UnidadesTableRegion(data).Render(r.Context(), &body)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#unidades-table-region"), datastar.WithModeOuter())
	_ = sse.MarshalAndPatchSignals(map[string]bool{"unidadSaving": false, "unidadSaved": true, "unidadesLoading": false})
}

func unidadFormError(message string) string {
	return `<div id="unidad-form-error" role="alert" class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">` + html.EscapeString(message) + `</div>`
}
