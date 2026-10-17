package roles

import (
	"database/sql"
	"encoding/json"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
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
	Detalles map[string]*model.Rol
	Pagina   int32
	Size     int32
	Error    string
}

type roleSignals struct {
	Pagina int32 `json:"pagina"`
	Size   int32 `json:"size"`
}

type roleFormSignals struct {
	Nombre      string   `json:"nombre"`
	Descripcion string   `json:"descripcion"`
	Jerarquia   int32    `json:"jerarquia"`
	Permisos    []string `json:"permisos"`
	Menus       []string `json:"menus"`
	Mode        string   `json:"rolMode"`
	ID          string   `json:"rolID"`
}

func (c *Controller) RolesRender(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		c.roleSave(w, r)
		return
	}
	data := c.rolesPageData(r)
	if r.Header.Get("Datastar-Request") == "true" {
		c.rolesPatch(w, r, data)
		return
	}
	_ = RenderRoles(data).Render(r.Context(), w)
}

func (c *Controller) rolesPageData(r *http.Request) RolesPageData {
	page := positiveInt(r.URL.Query().Get("pagina"), 1)
	size := positiveInt(r.URL.Query().Get("size"), 10)
	data := RolesPageData{Pagina: page, Size: size}
	data.Listado, _ = xrepo.Roles2(r.Context(), c.DB, model.QueryRoles{Pagina: page, Size: size})
	if data.Listado == nil {
		data.Error = "No fue posible cargar los roles."
	}
	data.Permisos, _ = xrepo.Permisos(r.Context(), c.DB)
	data.Menus, _ = xrepo.Menus(r.Context(), c.DB)
	data.Detalles = roleDetails(r, c.DB, data.Listado)
	return data
}

func roleDetails(r *http.Request, db *sql.DB, listado *model.RolPaginado) map[string]*model.Rol {
	details := make(map[string]*model.Rol)
	if listado == nil {
		return details
	}
	for _, role := range listado.Datos {
		if role == nil {
			continue
		}
		if detail, err := xrepo.RolByID(r.Context(), db, role.ID); err == nil && detail != nil {
			details[role.ID] = detail
		}
	}
	return details
}

func (c *Controller) roleSave(w http.ResponseWriter, r *http.Request) {
	var signals roleFormSignals
	if err := datastar.ReadSignals(r, &signals); err != nil {
		_ = r.ParseForm()
		signals.Nombre = r.FormValue("nombre")
		signals.Descripcion = r.FormValue("descripcion")
		signals.Mode = r.FormValue("rolMode")
		signals.ID = r.FormValue("rolID")
	}
	if signals.Mode == "" {
		signals.Mode = r.URL.Query().Get("mode")
	}
	if signals.ID == "" {
		signals.ID = r.URL.Query().Get("id")
	}
	input := model.NewRol{
		Nombre:      strings.TrimSpace(signals.Nombre),
		Descripcion: optionalString(signals.Descripcion),
		Jerarquia:   optionalInt(signals.Jerarquia),
		Permisos:    signals.Permisos,
		Menus:       roleMenuIDs(signals.Menus),
	}
	var err error
	if signals.Mode == "edit" && strings.TrimSpace(signals.ID) != "" {
		_, err = xrepo.UpdateRol(r.Context(), c.DB, model.UpdateRol{ID: signals.ID, Nombre: input.Nombre, Descripcion: input.Descripcion, Jerarquia: input.Jerarquia, Permisos: input.Permisos, Menus: input.Menus})
	} else {
		_, err = xrepo.CreateRol(r.Context(), c.DB, input)
	}
	sse := datastar.NewSSE(w, r)
	if err != nil {
		_ = sse.PatchElements(roleFormError(err.Error()), datastar.WithSelector("#rol-form-error"), datastar.WithModeOuter())
		_ = sse.MarshalAndPatchSignals(map[string]bool{"rolSaving": false, "rolSaved": false})
		return
	}
	data := c.rolesPageData(r)
	var body strings.Builder
	_ = RolesTableRegion(data).Render(r.Context(), &body)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#roles-table-region"), datastar.WithModeOuter())
	_ = sse.MarshalAndPatchSignals(map[string]bool{"rolSaving": false, "rolSaved": true})
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
		data.Detalles = roleDetails(r, c.DB, data.Listado)
	}
	var body strings.Builder
	_ = RolesTableRegion(data).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#roles-table-region"), datastar.WithModeOuter())
	_ = sse.ExecuteScript("history.replaceState({},''," + strconv.Quote(roleQueryURL(data)) + ")")
}

func roleEditAttributes(role *model.Rol) templ.Attributes {
	permissions := make([]string, 0, len(role.Permisos))
	for _, permission := range role.Permisos {
		if permission != nil {
			permissions = append(permissions, permission.Metodo)
		}
	}
	menus := make([]string, 0, len(role.Menus))
	for _, menu := range role.Menus {
		if menu != nil {
			menus = append(menus, menu.ID)
		}
	}
	encode := func(value any) string { data, _ := json.Marshal(value); return string(data) }
	sync := roleSelectionSyncScript()
	return templ.Attributes{
		"type": "button", "data-slot": "dialog-trigger", "data-base-ui-click-trigger": true,
		"data-templ-controls": "rol-dialog", "aria-haspopup": "dialog", "aria-expanded": "false",
		"data-on:click": "$rolMode='edit'; $rolID=" + encode(role.ID) + "; $rolSaving=false; $rolSaved=false; $nombre=" + encode(role.Nombre) + "; $descripcion=" + encode(optionalStringValue(role.Descripcion)) + "; $jerarquia=" + strconv.Itoa(int(role.Jerarquia)) + "; $permisos=" + encode(permissions) + "; $menus=" + encode(menus) + "; const root=document.getElementById('rol-dialog'); root?.querySelector('[data-rol-title]')?.replaceChildren('Editar rol'); root?.querySelector('#rol-form-error')?.replaceChildren(); " + sync + " syncRoleSelections('rol-permisos', " + encode(permissions) + "); syncRoleSelections('rol-menus', " + encode(menus) + ");",
	}
}

func roleCreateResetScript() string {
	return "$rolMode='create'; $rolID=''; $rolSaving=false; $rolSaved=false; $nombre=''; $descripcion=''; $jerarquia=0; $permisos=[]; $menus=[]; const root=document.getElementById('rol-dialog'); root?.querySelector('[data-rol-title]')?.replaceChildren('Nuevo rol'); root?.querySelector('#rol-form-error')?.replaceChildren(); " + roleSelectionSyncScript() + " syncRoleSelections('rol-permisos', []); syncRoleSelections('rol-menus', []);"
}

func roleSelectionSyncScript() string {
	return `const syncRoleSelections=(id,values)=>{const content=document.getElementById(id);if(!content)return;const wanted=new Set(values||[]);content.querySelectorAll('[data-slot="combobox-item"]').forEach(item=>{const selected=item.hasAttribute('data-selected');const shouldSelect=wanted.has(item.getAttribute('data-templ-value')||'');if(selected!==shouldSelect)item.click();});};`
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

func roleMenuIDs(values []string) []int32 {
	menus := make([]int32, 0, len(values))
	for _, value := range values {
		id, err := strconv.Atoi(value)
		if err == nil {
			menus = append(menus, int32(id))
		}
	}
	return menus
}

func roleDate(value time.Time) string {
	if value.IsZero() {
		return "—"
	}
	return value.Local().Format("02/01/2006 15:04")
}

func roleText(value *string) string { return optionalStringValue(value) }

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

func rolePageClick(page int32) string {
	if page < 1 {
		page = 1
	}
	return "$pagina=" + strconv.Itoa(int(page)) + "; @get('" + ADMIN_ROLES + "')"
}

func roleFormError(message string) string {
	return `<div id="rol-form-error" role="alert" class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">` + html.EscapeString(message) + `</div>`
}

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
