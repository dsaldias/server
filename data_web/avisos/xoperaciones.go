package avisos

import (
	"net/http"
	"strings"
	"time"

	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type avisoFormSignals struct {
	Mensaje  string `json:"avisoMensaje"`
	Desde    string `json:"avisoDesde"`
	Hasta    string `json:"avisoHasta"`
	Timezone string `json:"avisoTimezone"`
	ID       string `json:"avisoID"`
}

func (c *Controller) AvisoCrear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	c.avisoSave(w, r, false)
}

func (c *Controller) AvisoEditar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	c.avisoSave(w, r, true)
}

func (c *Controller) avisoSave(w http.ResponseWriter, r *http.Request, editing bool) {
	var signals avisoFormSignals
	if err := datastar.ReadSignals(r, &signals); err != nil {
		_ = r.ParseForm()
		signals.Mensaje = r.FormValue("mensaje")
		signals.Desde = r.FormValue("desde")
		signals.Hasta = r.FormValue("hasta")
		signals.Timezone = r.FormValue("avisoTimezone")
		signals.ID = r.FormValue("avisoID")
	}
	if strings.TrimSpace(signals.ID) == "" {
		signals.ID = strings.TrimSpace(r.URL.Query().Get("avisoid"))
	}
	desde, hasta, err := parseAvisoTimes(signals.Desde, signals.Hasta, signals.Timezone)
	if err == nil && strings.TrimSpace(signals.Mensaje) == "" {
		err = avisoValidationError("El mensaje es obligatorio.")
	}
	if err == nil && !hasta.After(desde) {
		err = avisoValidationError("La fecha final debe ser posterior a la fecha inicial.")
	}
	if err == nil {
		if editing {
			if signals.ID == "" {
				err = avisoValidationError("El identificador del aviso es obligatorio para editar.")
			} else {
				_, err = xrepo.UpdateNotificacion(r.Context(), c.DB, model.UpdNotificacion{ID: signals.ID, Mensaje: strings.TrimSpace(signals.Mensaje), Desde: desde, Hasta: hasta})
			}
		} else {
			_, err = xrepo.CrearNotificacion(r.Context(), c.DB, model.NewNotificacion{Mensaje: strings.TrimSpace(signals.Mensaje), Desde: desde, Hasta: hasta})
		}
	}
	sse := datastar.NewSSE(w, r)
	if err != nil {
		_ = sse.PatchElements(avisoFormError(err.Error()), datastar.WithSelector("#aviso-form-error"), datastar.WithModeOuter())
		_ = sse.MarshalAndPatchSignals(map[string]bool{"avisoSaving": false, "avisosLoading": false})
		return
	}
	data := c.avisosTableData(r)
	var body strings.Builder
	_ = AvisosTableRegion(data).Render(r.Context(), &body)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#avisos-table-region"), datastar.WithModeOuter())
	_ = sse.MarshalAndPatchSignals(map[string]bool{"avisoSaving": false, "avisosLoading": false})
	_ = sse.ExecuteScript("window.templ?.dialog?.close('aviso-dialog');")
}

type avisoValidationError string

func (e avisoValidationError) Error() string { return string(e) }

func parseAvisoTimes(desdeValue, hastaValue, timezone string) (time.Time, time.Time, error) {
	const layout = "2006-01-02T15:04"
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		timezone = "UTC"
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, time.Time{}, avisoValidationError("La zona horaria no es válida.")
	}
	desde, err := time.ParseInLocation(layout, strings.TrimSpace(desdeValue), location)
	if err != nil {
		return time.Time{}, time.Time{}, avisoValidationError("La fecha inicial no es válida.")
	}
	hasta, err := time.ParseInLocation(layout, strings.TrimSpace(hastaValue), location)
	if err != nil {
		return time.Time{}, time.Time{}, avisoValidationError("La fecha final no es válida.")
	}
	// Los avisos conservan la hora civil escrita por el usuario.
	desde = time.Date(desde.Year(), desde.Month(), desde.Day(), desde.Hour(), desde.Minute(), 0, 0, time.UTC)
	hasta = time.Date(hasta.Year(), hasta.Month(), hasta.Day(), hasta.Hour(), hasta.Minute(), 0, 0, time.UTC)
	return desde, hasta, nil
}
