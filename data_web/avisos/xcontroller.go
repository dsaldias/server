package avisos

import (
	"database/sql"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type Controller struct{ DB *sql.DB }

type AvisosPageData struct {
	Listado  []*model.Notificacion
	Detalles map[string]*model.Notificacion
	Error    string
}

type avisoFormSignals struct {
	Mensaje       string
	Desde         string
	Hasta         string
	AvisoMode     string
	AvisoID       string
	AvisoTimezone string
}

func (c *Controller) AvisosRender(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		c.avisoSave(w, r)
		return
	}
	data := c.avisosPageData(r)
	if r.Header.Get("Datastar-Request") == "true" {
		c.avisosPatch(w, r, data)
		return
	}
	_ = RenderAvisos(data).Render(r.Context(), w)
}

func (c *Controller) avisosPageData(r *http.Request) AvisosPageData {
	data := AvisosPageData{Detalles: make(map[string]*model.Notificacion)}
	listado, err := xrepo.Notificaciones(r.Context(), c.DB)
	if err != nil {
		data.Error = "No fue posible cargar los avisos activos."
		return data
	}
	data.Listado = listado
	for _, aviso := range listado {
		if aviso != nil {
			data.Detalles[aviso.ID] = aviso
		}
	}
	return data
}

func (c *Controller) avisosPatch(w http.ResponseWriter, r *http.Request, data AvisosPageData) {
	data = c.avisosPageData(r)
	var body strings.Builder
	_ = AvisosTableRegion(data).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#avisos-table-region"), datastar.WithModeOuter())
}

func (c *Controller) avisoSave(w http.ResponseWriter, r *http.Request) {
	var signals avisoFormSignals
	if err := datastar.ReadSignals(r, &signals); err != nil {
		_ = r.ParseForm()
		signals.Mensaje = r.FormValue("mensaje")
		signals.Desde = r.FormValue("desde")
		signals.Hasta = r.FormValue("hasta")
		signals.AvisoMode = r.FormValue("avisoMode")
		signals.AvisoID = r.FormValue("avisoID")
		signals.AvisoTimezone = r.FormValue("avisoTimezone")
	}
	desde, hasta, err := parseAvisoTimes(signals.Desde, signals.Hasta, signals.AvisoTimezone)
	if err == nil && strings.TrimSpace(signals.Mensaje) == "" {
		err = avisoValidationError("El mensaje es obligatorio.")
	}
	if err == nil && !hasta.After(desde) {
		err = avisoValidationError("La fecha final debe ser posterior a la fecha inicial.")
	}
	if err == nil {
		if signals.AvisoMode == "edit" && strings.TrimSpace(signals.AvisoID) != "" {
			_, err = xrepo.UpdateNotificacion(r.Context(), c.DB, model.UpdNotificacion{ID: signals.AvisoID, Mensaje: strings.TrimSpace(signals.Mensaje), Desde: desde, Hasta: hasta})
		} else {
			_, err = xrepo.CrearNotificacion(r.Context(), c.DB, model.NewNotificacion{Mensaje: strings.TrimSpace(signals.Mensaje), Desde: desde, Hasta: hasta})
		}
	}
	sse := datastar.NewSSE(w, r)
	if err != nil {
		_ = sse.PatchElements(avisoFormError(err.Error()), datastar.WithSelector("#aviso-form-error"), datastar.WithModeOuter())
		_ = sse.MarshalAndPatchSignals(map[string]bool{"avisoSaving": false, "avisoSaved": false})
		return
	}
	data := c.avisosPageData(r)
	var body strings.Builder
	_ = AvisosTableRegion(data).Render(r.Context(), &body)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#avisos-table-region"), datastar.WithModeOuter())
	_ = sse.MarshalAndPatchSignals(map[string]bool{"avisoSaving": false, "avisoSaved": true})
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
	// Los avisos usan DATETIME como hora civil: se conservan los componentes
	// escritos por el usuario y no se convierten a UTC.
	desde = time.Date(desde.Year(), desde.Month(), desde.Day(), desde.Hour(), desde.Minute(), 0, 0, time.UTC)
	hasta = time.Date(hasta.Year(), hasta.Month(), hasta.Day(), hasta.Hour(), hasta.Minute(), 0, 0, time.UTC)
	return desde, hasta, nil
}

func avisoDate(value time.Time) string {
	if value.IsZero() {
		return "—"
	}
	return value.UTC().Format("02/01/2006 15:04")
}

func avisoDateISO(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("2006-01-02T15:04")
}

func avisoDateInputExpression(value time.Time) string {
	iso := avisoDateISO(value)
	if iso == "" {
		return "''"
	}
	return strconv.Quote(iso)
}

func avisoSignalString(value string) string { return strconv.Quote(value) }

func avisoEditAttributes(aviso *model.Notificacion) templ.Attributes {
	return templ.Attributes{
		"type":                       "button",
		"data-slot":                  "dialog-trigger",
		"data-base-ui-click-trigger": true,
		"data-templ-controls":        "aviso-dialog",
		"aria-haspopup":              "dialog",
		"aria-expanded":              "false",
		"data-on:click":              "$avisoMode='edit'; $avisoID=" + strconv.Quote(aviso.ID) + "; $avisoSaving=false; $avisoSaved=false; $avisoTimezone=Intl.DateTimeFormat().resolvedOptions().timeZone; $mensaje=" + strconv.Quote(aviso.Mensaje) + "; $desde=" + avisoDateInputExpression(aviso.Desde) + "; $hasta=" + avisoDateInputExpression(aviso.Hasta) + "; document.getElementById('aviso-form-error')?.replaceChildren();",
	}
}

func avisoViewAttributes(id string) templ.Attributes {
	return templ.Attributes{"type": "button", "data-slot": "dialog-trigger", "data-base-ui-click-trigger": true, "data-templ-controls": "aviso-view-" + id, "aria-haspopup": "dialog", "aria-expanded": "false"}
}

func avisoCreateResetScript() string {
	return "$avisoMode='create'; $avisoID=''; $avisoSaving=false; $avisoSaved=false; $mensaje=''; $desde=''; $hasta=''; document.getElementById('aviso-form-error')?.replaceChildren();"
}

func avisoFormError(message string) string {
	return "<div id=\"aviso-form-error\" role=\"alert\" class=\"rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive\">" + html.EscapeString(message) + "</div>"
}
