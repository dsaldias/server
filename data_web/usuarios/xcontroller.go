package usuarios

import (
	"database/sql"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dsaldias/server/data_web/components/combobox"
	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type Controller struct {
	DB *sql.DB
}

type UsuariosPageData struct {
	Listado *model.UsuarioPaginado
	Roles   []*model.ResponseRoles
	Pagina  int32
	Size    int32
	Rol     string
	Nombre  string
	Error   string
}

type usuariosSignals struct {
	Pagina int32  `json:"pagina"`
	Size   int32  `json:"size"`
	Rol    string `json:"rol"`
	Nombre string `json:"nombre"`
}

func (c *Controller) UsuariosRender(w http.ResponseWriter, r *http.Request) {
	data := c.usuariosPageData(r)
	_ = RenderUsuarios(data).Render(r.Context(), w)
}

func (c *Controller) UsuariosListaHTML(w http.ResponseWriter, r *http.Request) {
	data := c.usuariosTableData(r)
	if r.Header.Get("Datastar-Request") == "true" {
		c.usuariosPatch(w, r, data)
		return
	}
	_ = UsuariosTableRegion(data).Render(r.Context(), w)
}

func (c *Controller) usuariosPageData(r *http.Request) UsuariosPageData {
	data := c.usuariosTableData(r)
	data.Roles, _ = xrepo.Roles(r.Context(), c.DB)
	return data
}

func (c *Controller) usuariosTableData(r *http.Request) UsuariosPageData {
	page := parsePositiveInt(r.URL.Query().Get("pagina"), 1)
	size := parsePositiveInt(r.URL.Query().Get("size"), 10)
	rol := strings.TrimSpace(r.URL.Query().Get("rol"))
	nombre := strings.TrimSpace(r.URL.Query().Get("nombre"))
	data := UsuariosPageData{Pagina: page, Size: size, Rol: rol, Nombre: nombre}
	query := model.QueryUsuarios{Pagina: page, Size: size}
	if rol != "" {
		query.Rol = &rol
	}
	if nombre != "" {
		query.Nombre = &nombre
	}
	data.Listado, _ = xrepo.Usuarios2(r.Context(), c.DB, query)
	if data.Listado == nil {
		data.Error = "No fue posible cargar los usuarios."
	}
	return data
}

func (c *Controller) usuariosPatch(w http.ResponseWriter, r *http.Request, data UsuariosPageData) {
	var signals usuariosSignals
	if err := datastar.ReadSignals(r, &signals); err == nil {
		if signals.Pagina > 0 {
			data.Pagina = signals.Pagina
		}
		if signals.Size > 0 {
			data.Size = signals.Size
		}
		data.Rol = strings.TrimSpace(signals.Rol)
		data.Nombre = strings.TrimSpace(signals.Nombre)
		query := model.QueryUsuarios{Pagina: data.Pagina, Size: data.Size}
		if data.Rol != "" {
			query.Rol = &data.Rol
		}
		if data.Nombre != "" {
			query.Nombre = &data.Nombre
		}
		data.Listado, _ = xrepo.Usuarios2(r.Context(), c.DB, query)
		if data.Listado == nil {
			data.Error = "No fue posible cargar los usuarios."
		}
	}
	var body strings.Builder
	_ = UsuariosTableRegion(data).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#usuarios-table-region"), datastar.WithModeOuter())
	_ = sse.ExecuteScript("history.replaceState({},''," + strconv.Quote(usuarioQueryURL(data)) + ")")
	_ = sse.MarshalAndPatchSignals(map[string]bool{"usuariosLoading": false})
}

func usuarioQueryURL(data UsuariosPageData) string {
	query := "?pagina=" + strconv.Itoa(int(data.Pagina)) + "&size=" + strconv.Itoa(int(data.Size))
	if data.Rol != "" {
		query += "&rol=" + urlQueryEscape(data.Rol)
	}
	if data.Nombre != "" {
		query += "&nombre=" + urlQueryEscape(data.Nombre)
	}
	return ADMIN_USUARIOS + query
}

func urlQueryEscape(value string) string { return url.QueryEscape(value) }

func parsePositiveInt(value string, fallback int32) int32 {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return int32(parsed)
}

func usuarioSignalString(value string) string { return strconv.Quote(value) }

func usuarioSignalInt(value int32) string { return strconv.Itoa(int(value)) }

func usuariosRoleItems(roles []*model.ResponseRoles) []combobox.ItemData {
	items := make([]combobox.ItemData, 0, len(roles))
	for _, role := range roles {
		if role != nil {
			items = append(items, combobox.ItemData{Value: role.ID, Label: role.Nombre})
		}
	}
	return items
}

func optionalString(value *string) string {
	if value == nil {
		return "—"
	}
	return *value
}

func optionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func usuarioStatusLabel(active bool) string {
	if active {
		return "Activo"
	}
	return "Inactivo"
}

func usuarioStatusClass(active bool) string {
	if active {
		return "inline-flex rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300"
	}
	return "inline-flex rounded-full bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground"
}

func usuarioLastLogin(value *time.Time) string {
	if value == nil || value.IsZero() {
		return "Nunca"
	}
	return value.Local().Format("02/01/2006 15:04")
}

func usuarioTime(value time.Time) string {
	if value.IsZero() {
		return "—"
	}
	return value.Local().Format("02/01/2006 15:04")
}

func usuarioFloat(value *float64) string {
	if value == nil {
		return "—"
	}
	return strconv.FormatFloat(*value, 'f', 6, 64)
}

func usuarioInt(value int32) string { return strconv.Itoa(int(value)) }

func usuarioPaginationSummary(data *model.UsuarioPaginado) string {
	if data == nil || data.Paginacion == nil {
		return "Sin registros"
	}
	return strconv.Itoa(int(data.Paginacion.Registros)) + " registros"
}

func usuarioPageLabel(data *model.UsuarioPaginado) string {
	if data == nil || data.Paginacion == nil {
		return "1 / 1"
	}
	return strconv.Itoa(int(data.Paginacion.Pagina)) + " / " + strconv.Itoa(int(data.Paginacion.Paginas))
}

func usuarioCanPrevious(data *model.UsuarioPaginado) bool {
	return data != nil && data.Paginacion != nil && data.Paginacion.Pagina > 1
}

func usuarioCanNext(data *model.UsuarioPaginado) bool {
	return data != nil && data.Paginacion != nil && data.Paginacion.Pagina < data.Paginacion.Paginas
}

func usuarioPageClick(page int32) string {
	return "const button = evt.currentTarget; if (button) { button.disabled = true; button.setAttribute('aria-busy','true'); button.innerHTML = '<span class=\"inline-flex items-center gap-2\"><span class=\"size-3.5 animate-spin rounded-full border-2 border-current border-t-transparent\"></span>Cargando...</span>'; } $pagina = " + strconv.Itoa(int(page)) + "; @get('" + ADMIN_USUARIOS_LISTAHTML + "')"
}

func usuarioRefreshScript() string {
	return "$usuariosLoading = true; $pagina = 1; @get('" + ADMIN_USUARIOS_LISTAHTML + "')"
}

func usuarioLoadingButtonEffectScript(buttonSelector, iconSelector, labelSelector, label string) string {
	return "const button = document.querySelector('" + buttonSelector + "'); if (button) { button.disabled = $usuariosLoading; if ($usuariosLoading) { button.setAttribute('aria-busy','true'); button.querySelector('" + iconSelector + "')?.classList.add('hidden'); const label = button.querySelector('" + labelSelector + "'); if (label) { label.classList.remove('sr-only'); label.textContent = 'Cargando...'; } } else { button.removeAttribute('aria-busy'); button.querySelector('" + iconSelector + "')?.classList.remove('hidden'); const label = button.querySelector('" + labelSelector + "'); if (label) { label.classList.add('sr-only'); label.textContent = '" + label + "'; } } }"
}

func usuarioRoleFilterScript() string {
	return "$rol = evt.detail.values[0]; $pagina = 1; @get('" + ADMIN_USUARIOS_LISTAHTML + "')"
}

func usuarioPageSizeScript() string {
	return "$pagina = 1; @get('" + ADMIN_USUARIOS_LISTAHTML + "')"
}

func usuarioNameSearchScript() string {
	return "$usuariosLoading = true; $pagina = 1; @get('" + ADMIN_USUARIOS_LISTAHTML + "')"
}

func usuarioNameSearchKeydownScript() string {
	return "if (evt.key === 'Enter') { $usuariosLoading = true; $pagina = 1; @get('" + ADMIN_USUARIOS_LISTAHTML + "') }"
}
