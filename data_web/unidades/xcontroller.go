package unidades

import (
	"database/sql"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type Controller struct{ DB *sql.DB }

type UnidadesPageData struct {
	Listado *model.UnidadPaginada
	Pagina  int32
	Size    int32
	Nombre  string
	Error   string
}

type unidadSignals struct {
	Pagina int32  `json:"pagina"`
	Size   int32  `json:"size"`
	Nombre string `json:"nombre"`
}

func (c *Controller) UnidadesRender(w http.ResponseWriter, r *http.Request) {
	data := c.unidadesTableData(r)
	_ = RenderUnidades(data).Render(r.Context(), w)
}

func (c *Controller) UnidadesListaHTML(w http.ResponseWriter, r *http.Request) {
	data := c.unidadesTableData(r)
	if r.Header.Get("Datastar-Request") == "true" {
		var signals unidadSignals
		if err := datastar.ReadSignals(r, &signals); err == nil {
			if signals.Pagina > 0 {
				data.Pagina = signals.Pagina
			}
			if signals.Size > 0 {
				data.Size = signals.Size
			}
			data.Nombre = strings.TrimSpace(signals.Nombre)
			data.Listado, _ = xrepo.Unidades2(r.Context(), c.DB, model.QueryUnidades{Pagina: data.Pagina, Size: data.Size, Nombre: optionalString(data.Nombre)})
			if data.Listado == nil {
				data.Error = "No fue posible cargar las unidades."
			}
		}
		c.unidadesPatch(w, r, data)
		return
	}
	_ = UnidadesTableRegion(data).Render(r.Context(), w)
}

func (c *Controller) unidadesTableData(r *http.Request) UnidadesPageData {
	page := positiveInt(r.URL.Query().Get("pagina"), 1)
	size := positiveInt(r.URL.Query().Get("size"), 10)
	nombre := strings.TrimSpace(r.URL.Query().Get("nombre"))
	data := UnidadesPageData{Pagina: page, Size: size, Nombre: nombre}
	data.Listado, _ = xrepo.Unidades2(r.Context(), c.DB, model.QueryUnidades{Pagina: page, Size: size, Nombre: optionalString(nombre)})
	if data.Listado == nil {
		data.Error = "No fue posible cargar las unidades."
	}
	return data
}

func (c *Controller) unidadesPatch(w http.ResponseWriter, r *http.Request, data UnidadesPageData) {
	var body strings.Builder
	_ = UnidadesTableRegion(data).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#unidades-table-region"), datastar.WithModeOuter())
	_ = sse.ExecuteScript("history.replaceState({},''," + strconv.Quote(unidadQueryURL(data)) + ")")
	_ = sse.MarshalAndPatchSignals(map[string]any{"unidadesLoading": false})
}

func unidadRefreshScript() string {
	return "$unidadesLoading=true; $pagina=1; @get('" + ADMIN_UNIDADES_LISTAHTML + "')"
}

func unidadSearchScript() string {
	return "$unidadesLoading=true; $pagina=1; @get('" + ADMIN_UNIDADES_LISTAHTML + "')"
}

func unidadSearchKeydownScript() string {
	return "if(evt.key==='Enter'){" + unidadSearchScript() + "}"
}

func unidadPageSizeScript() string {
	return "$unidadesLoading=true; $pagina=1; @get('" + ADMIN_UNIDADES_LISTAHTML + "')"
}

func unidadPageClick(page int32) string {
	if page < 1 {
		page = 1
	}
	return "$pagina=" + strconv.Itoa(int(page)) + ";$unidadesLoading=true; @get('" + ADMIN_UNIDADES_LISTAHTML + "')"
}

func unidadLoadingButtonEffectScript() string {
	return "const button=document.querySelector('[data-unidades-refresh]');if(button){button.disabled=$unidadesLoading;if($unidadesLoading){button.setAttribute('aria-busy','true');button.querySelector('[data-unidades-refresh-icon]')?.classList.add('hidden');const label=button.querySelector('[data-unidades-refresh-label]');if(label)label.textContent='Cargando...';}else{button.removeAttribute('aria-busy');button.querySelector('[data-unidades-refresh-icon]')?.classList.remove('hidden');const label=button.querySelector('[data-unidades-refresh-label]');if(label)label.textContent='Refrescar';}}"
}

func unidadSearchLoadingButtonEffectScript() string {
	return "const button=document.querySelector('[data-unidades-search]');if(button){button.disabled=$unidadesLoading;if($unidadesLoading){button.setAttribute('aria-busy','true');button.querySelector('[data-unidades-search-icon]')?.classList.add('hidden');const label=button.querySelector('[data-unidades-search-label]');if(label){label.classList.remove('sr-only');label.textContent='Cargando...';}}else{button.removeAttribute('aria-busy');button.querySelector('[data-unidades-search-icon]')?.classList.remove('hidden');const label=button.querySelector('[data-unidades-search-label]');if(label){label.classList.add('sr-only');label.textContent='Buscar';}}}"
}

func positiveInt(value string, fallback int32) int32 {
	number, err := strconv.Atoi(value)
	if err != nil || number < 1 {
		return fallback
	}
	return int32(number)
}

func unidadQueryURL(data UnidadesPageData) string {
	query := ADMIN_UNIDADES + "?pagina=" + strconv.Itoa(int(data.Pagina)) + "&size=" + strconv.Itoa(int(data.Size))
	if data.Nombre != "" {
		query += "&nombre=" + url.QueryEscape(data.Nombre)
	}
	return query
}

func unidadSignalInt(value int32) string     { return strconv.Itoa(int(value)) }
func unidadSignalString(value string) string { return strconv.Quote(value) }
func unidadStatus(value int32) string        { return strconv.Itoa(int(value)) }

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
	return strconv.Itoa(len(data.Datos)) + " de " + strconv.Itoa(int(data.Paginacion.Registros)) + " registros"
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
