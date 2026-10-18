package roles

import (
	"database/sql"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dsaldias/server/data_web/components/combobox"
	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type Controller struct{ DB *sql.DB }

type RolesPageData struct {
	Listado  *model.RolPaginado
	Permisos []*model.Permiso
	Menus    []*model.Menus
	Pagina   int32
	Size     int32
	Error    string
}

type roleSignals struct {
	Pagina int32 `json:"pagina"`
	Size   int32 `json:"size"`
}

func (c *Controller) RolesRender(w http.ResponseWriter, r *http.Request) {
	data := c.rolesPageData(r)
	_ = RenderRoles(data).Render(r.Context(), w)
}

func (c *Controller) RolesListaHTML(w http.ResponseWriter, r *http.Request) {
	data := c.rolesTableData(r)
	if r.Header.Get("Datastar-Request") == "true" {
		c.rolesPatch(w, r, data)
		return
	}
	_ = RolesTableRegion(data).Render(r.Context(), w)
}

func (c *Controller) rolesPageData(r *http.Request) RolesPageData {
	return c.rolesTableData(r)
}

func (c *Controller) rolesTableData(r *http.Request) RolesPageData {
	page := positiveInt(r.URL.Query().Get("pagina"), 1)
	size := positiveInt(r.URL.Query().Get("size"), 10)
	data := RolesPageData{Pagina: page, Size: size}
	data.Listado, _ = xrepo.Roles2(r.Context(), c.DB, model.QueryRoles{Pagina: page, Size: size})
	if data.Listado == nil {
		data.Error = "No fue posible cargar los roles."
	}
	return data
}

func (c *Controller) rolesPatch(w http.ResponseWriter, r *http.Request, data RolesPageData) {
	var signals roleSignals
	if err := datastar.ReadSignals(r, &signals); err == nil {
		if signals.Pagina > 0 {
			data.Pagina = signals.Pagina
		}
		if signals.Size > 0 {
			data.Size = signals.Size
		}
		data.Listado, _ = xrepo.Roles2(r.Context(), c.DB, model.QueryRoles{Pagina: data.Pagina, Size: data.Size})
		if data.Listado == nil {
			data.Error = "No fue posible cargar los roles."
		}
	}
	var body strings.Builder
	_ = RolesTableRegion(data).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#roles-table-region"), datastar.WithModeOuter())
	_ = sse.ExecuteScript("history.replaceState({},''," + strconv.Quote(roleQueryURL(data)) + ")")
	_ = sse.MarshalAndPatchSignals(map[string]bool{"rolesLoading": false})
}

func positiveInt(value string, fallback int32) int32 {
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return fallback
	}
	return int32(n)
}

func roleQueryURL(data RolesPageData) string {
	return ADMIN_ROLES + "?pagina=" + strconv.Itoa(int(data.Pagina)) + "&size=" + strconv.Itoa(int(data.Size))
}

func roleSignalString(value string) string { return strconv.Quote(value) }
func roleSignalInt(value int32) string     { return strconv.Itoa(int(value)) }
func roleStatus(value int32) string        { return strconv.Itoa(int(value)) }

func rolePageClick(page int32) string {
	if page < 1 {
		page = 1
	}
	return "const button=evt.currentTarget;if(button){button.disabled=true;button.setAttribute('aria-busy','true');button.innerHTML='<span class=\"inline-flex items-center gap-2\"><span class=\"size-3.5 animate-spin rounded-full border-2 border-current border-t-transparent\"></span>Cargando...</span>';}$pagina=" + strconv.Itoa(int(page)) + "; @get('" + ADMIN_ROLES_LISTAHTML + "')"
}

func roleRefreshScript() string {
	return "$rolesLoading=true; $pagina=1; @get('" + ADMIN_ROLES_LISTAHTML + "')"
}

func rolePageSizeScript() string {
	return "$pagina=1; @get('" + ADMIN_ROLES_LISTAHTML + "')"
}

func roleLoadingButtonEffectScript() string {
	return "const button=document.querySelector('[data-roles-refresh]');if(button){button.disabled=$rolesLoading;if($rolesLoading){button.setAttribute('aria-busy','true');button.querySelector('[data-roles-refresh-icon]')?.classList.add('hidden');const label=button.querySelector('[data-roles-refresh-label]');if(label)label.textContent='Cargando...';}else{button.removeAttribute('aria-busy');button.querySelector('[data-roles-refresh-icon]')?.classList.remove('hidden');const label=button.querySelector('[data-roles-refresh-label]');if(label)label.textContent='Refrescar';}}"
}

func rolePageSummary(data *model.RolPaginado) string {
	if data == nil || data.Paginacion == nil {
		return "Sin registros"
	}
	return "Mostrando " + strconv.Itoa(len(data.Datos)) + " de " + strconv.Itoa(int(data.Paginacion.Registros)) + " registros"
}

func rolePageLabel(data *model.RolPaginado) string {
	if data == nil || data.Paginacion == nil {
		return "—"
	}
	return "Página " + strconv.Itoa(int(data.Paginacion.Pagina)) + " de " + strconv.Itoa(int(data.Paginacion.Paginas))
}

func roleCanPrevious(data *model.RolPaginado) bool {
	return data != nil && data.Paginacion != nil && data.Paginacion.Pagina > 1
}

func roleCanNext(data *model.RolPaginado) bool {
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

func optionalInt(value int32) *int32 {
	if value == 0 {
		return nil
	}
	return &value
}

func roleDate(value time.Time) string {
	if value.IsZero() {
		return "—"
	}
	return value.Local().Format("02/01/2006 15:04")
}

func roleText(value *string) string { return optionalStringValue(value) }

func rolePermissionItems(values []*model.Permiso) []combobox.ItemData {
	items := make([]combobox.ItemData, 0, len(values))
	for _, value := range values {
		if value != nil {
			items = append(items, combobox.ItemData{Value: value.Metodo, Label: value.Nombre})
		}
	}
	return items
}

func roleMenuItems(values []*model.Menus) []combobox.ItemData {
	items := make([]combobox.ItemData, 0, len(values))
	for _, value := range values {
		if value != nil {
			items = append(items, combobox.ItemData{Value: value.ID, Label: value.Label})
		}
	}
	return items
}

func roleMenuParentName(menu *model.Menus, menus []*model.Menus) string {
	if menu == nil || menu.PadreID == nil {
		return ""
	}
	for _, parent := range menus {
		if parent != nil && parent.ID == *menu.PadreID {
			return "de " + parent.Label
		}
	}
	return "de menú padre"
}

func roleMenuClass(menu *model.Menus) string {
	if menu != nil && menu.PadreID != nil {
		return "flex items-start gap-2 rounded-md border border-border/40 bg-muted/10 px-3 py-2 pl-6"
	}
	return "flex items-start gap-2 rounded-md border border-border/60 bg-muted/20 px-3 py-2"
}

func roleFormError(message string) string {
	return `<div id="rol-form-error" role="alert" class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">` + html.EscapeString(message) + `</div>`
}
