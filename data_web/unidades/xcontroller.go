package unidades

import (
	"database/sql"
	"encoding/json"
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

type UnidadesPageData struct {
	Listado  *model.UnidadPaginada
	Detalles map[string]*model.Unidad
	Pagina   int32
	Size     int32
	Error    string
}

type unidadSignals struct {
	Pagina int32 `json:"pagina"`
	Size   int32 `json:"size"`
}
type unidadFormSignals struct {
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Orden       int32  `json:"orden"`
	Latitud     string `json:"latitud"`
	Longitud    string `json:"longitud"`
	Mode        string `json:"unidadMode"`
	ID          string `json:"unidadID"`
}

func (c *Controller) UnidadesRender(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		c.unidadSave(w, r)
		return
	}
	data := c.unidadesPageData(r)
	if r.Header.Get("Datastar-Request") == "true" {
		c.unidadesPatch(w, r, data)
		return
	}
	_ = RenderUnidades(data).Render(r.Context(), w)
}

func (c *Controller) unidadesPageData(r *http.Request) UnidadesPageData {
	page, size := positiveInt(r.URL.Query().Get("pagina"), 1), positiveInt(r.URL.Query().Get("size"), 10)
	data := UnidadesPageData{Pagina: page, Size: size}
	data.Listado, _ = xrepo.Unidades2(r.Context(), c.DB, model.QueryUnidades{Pagina: page, Size: size})
	if data.Listado == nil {
		data.Error = "No fue posible cargar las unidades."
	}
	data.Detalles = unidadDetails(r, c.DB, data.Listado)
	return data
}

func unidadDetails(r *http.Request, db *sql.DB, listado *model.UnidadPaginada) map[string]*model.Unidad {
	details := make(map[string]*model.Unidad)
	if listado == nil {
		return details
	}
	for _, unidad := range listado.Datos {
		if unidad == nil {
			continue
		}
		if detail, err := xrepo.UnidadByID(r.Context(), db, unidad.ID); err == nil && detail != nil {
			details[unidad.ID] = detail
		}
	}
	return details
}

func (c *Controller) unidadSave(w http.ResponseWriter, r *http.Request) {
	var signals unidadFormSignals
	if err := datastar.ReadSignals(r, &signals); err != nil {
		_ = r.ParseForm()
		signals.Nombre, signals.Descripcion, signals.Latitud, signals.Longitud = r.FormValue("nombre"), r.FormValue("descripcion"), r.FormValue("latitud"), r.FormValue("longitud")
		signals.Mode, signals.ID = r.FormValue("unidadMode"), r.FormValue("unidadID")
	}
	if signals.Mode == "" {
		signals.Mode = r.URL.Query().Get("mode")
	}
	if signals.ID == "" {
		signals.ID = r.URL.Query().Get("id")
	}
	input := model.NewUnidad{Nombre: strings.TrimSpace(signals.Nombre), Descripcion: optionalString(signals.Descripcion), Orden: signals.Orden, Latitud: optionalFloat(signals.Latitud), Longitud: optionalFloat(signals.Longitud)}
	var err error
	if signals.Mode == "edit" && strings.TrimSpace(signals.ID) != "" {
		_, err = xrepo.UpdateUnidad(r.Context(), c.DB, model.UpdUnidad{ID: signals.ID, Nombre: input.Nombre, Descripcion: input.Descripcion, Orden: input.Orden, Latitud: input.Latitud, Longitud: input.Longitud})
	} else {
		_, err = xrepo.CreateUnidad(r.Context(), c.DB, input)
	}
	sse := datastar.NewSSE(w, r)
	if err != nil {
		_ = sse.PatchElements(unidadFormError(err.Error()), datastar.WithSelector("#unidad-form-error"), datastar.WithModeOuter())
		_ = sse.MarshalAndPatchSignals(map[string]bool{"unidadSaving": false, "unidadSaved": false})
		return
	}
	data := c.unidadesPageData(r)
	var body strings.Builder
	_ = UnidadesTableRegion(data).Render(r.Context(), &body)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#unidades-table-region"), datastar.WithModeOuter())
	_ = sse.MarshalAndPatchSignals(map[string]bool{"unidadSaving": false, "unidadSaved": true})
}

func (c *Controller) unidadesPatch(w http.ResponseWriter, r *http.Request, data UnidadesPageData) {
	var signals unidadSignals
	if err := datastar.ReadSignals(r, &signals); err == nil {
		if signals.Pagina > 0 {
			data.Pagina = signals.Pagina
		}
		if signals.Size > 0 {
			data.Size = signals.Size
		}
		data.Listado, _ = xrepo.Unidades2(r.Context(), c.DB, model.QueryUnidades{Pagina: data.Pagina, Size: data.Size})
		data.Detalles = unidadDetails(r, c.DB, data.Listado)
	}
	var body strings.Builder
	_ = UnidadesTableRegion(data).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#unidades-table-region"), datastar.WithModeOuter())
	_ = sse.ExecuteScript("history.replaceState({},''," + strconv.Quote(unidadQueryURL(data)) + ")")
}

func unidadEditAttributes(unidad *model.Unidad) templ.Attributes {
	encode := func(value any) string { data, _ := json.Marshal(value); return string(data) }
	return templ.Attributes{"type": "button", "data-slot": "dialog-trigger", "data-base-ui-click-trigger": true, "data-templ-controls": "unidad-dialog", "aria-haspopup": "dialog", "aria-expanded": "false", "data-on:click": "$unidadMode='edit'; $unidadID=" + encode(unidad.ID) + "; $unidadSaving=false; $unidadSaved=false; $nombre=" + encode(unidad.Nombre) + "; $descripcion=" + encode(optionalStringValue(unidad.Descripcion)) + "; $orden=" + strconv.Itoa(int(unidad.Orden)) + "; $latitud=" + encode(optionalFloatValue(unidad.Latitud)) + "; $longitud=" + encode(optionalFloatValue(unidad.Longitud)) + "; const root=document.getElementById('unidad-dialog'); root?.querySelector('[data-unidad-title]')?.replaceChildren('Editar unidad'); root?.querySelector('#unidad-form-error')?.replaceChildren();"}
}

func unidadCreateResetScript() string {
	return "$unidadMode='create'; $unidadID=''; $unidadSaving=false; $unidadSaved=false; $nombre=''; $descripcion=''; $orden=0; $latitud=''; $longitud=''; const root=document.getElementById('unidad-dialog'); root?.querySelector('[data-unidad-title]')?.replaceChildren('Nueva unidad'); root?.querySelector('#unidad-form-error')?.replaceChildren();"
}
func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
func optionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
func optionalFloat(value string) *float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}
	return &number
}
func optionalFloatValue(value *float64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatFloat(*value, 'f', 6, 64)
}
func positiveInt(value string, fallback int32) int32 {
	number, err := strconv.Atoi(value)
	if err != nil || number < 1 {
		return fallback
	}
	return int32(number)
}
func unidadQueryURL(data UnidadesPageData) string {
	return ADMIN_UNIDADES + "?pagina=" + strconv.Itoa(int(data.Pagina)) + "&size=" + strconv.Itoa(int(data.Size))
}
func unidadSignalInt(value int32) string { return strconv.Itoa(int(value)) }
func unidadStatus(value int32) string    { return strconv.Itoa(int(value)) }
func unidadDate(value time.Time) string {
	if value.IsZero() {
		return "—"
	}
	return value.Local().Format("02/01/2006 15:04")
}
func unidadText(value *string) string { return optionalStringValue(value) }
func unidadFloat(value *float64) string {
	if value == nil {
		return "—"
	}
	return strconv.FormatFloat(*value, 'f', 6, 64)
}
func unidadPageSummary(data *model.UnidadPaginada) string {
	if data == nil || data.Paginacion == nil {
		return "Sin registros"
	}
	return "Mostrando " + strconv.Itoa(len(data.Datos)) + " de " + strconv.Itoa(int(data.Paginacion.Registros)) + " registros"
}
func unidadPageLabel(data *model.UnidadPaginada) string {
	if data == nil || data.Paginacion == nil {
		return "—"
	}
	return "Página " + strconv.Itoa(int(data.Paginacion.Pagina)) + " de " + strconv.Itoa(int(data.Paginacion.Paginas))
}
func unidadCanPrevious(data *model.UnidadPaginada) bool {
	return data != nil && data.Paginacion != nil && data.Paginacion.Pagina > 1
}
func unidadCanNext(data *model.UnidadPaginada) bool {
	return data != nil && data.Paginacion != nil && data.Paginacion.Pagina < data.Paginacion.Paginas
}
func unidadPageClick(page int32) string {
	if page < 1 {
		page = 1
	}
	return "$pagina=" + strconv.Itoa(int(page)) + "; @get('" + ADMIN_UNIDADES + "')"
}
func unidadFormError(message string) string {
	return `<div id="unidad-form-error" role="alert" class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">` + html.EscapeString(message) + `</div>`
}
