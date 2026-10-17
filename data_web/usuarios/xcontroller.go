package usuarios

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/data_web/components/combobox"
	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type Controller struct {
	DB *sql.DB
}

type UsuariosPageData struct {
	Listado  *model.UsuarioPaginado
	Roles    []*model.ResponseRoles
	Unidades []*model.Unidad
	Detalles map[string]*model.ResponseUsuario
	Pagina   int32
	Size     int32
	Rol      string
	Error    string
}

type usuariosSignals struct {
	Pagina int32  `json:"pagina"`
	Size   int32  `json:"size"`
	Rol    string `json:"rol"`
}

type usuarioFormSignals struct {
	Nombres   string `json:"nombres"`
	Apellido1 string `json:"apellido1"`
	Apellido2 string `json:"apellido2"`
	Documento string `json:"documento"`
	Celular   string `json:"celular"`
	Correo    string `json:"correo"`
	Sexo      string `json:"sexo"`
	Direccion string `json:"direccion"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Foto64    string `json:"foto64"`
	RolesJSON string `json:"rolesJson"`
	Mode      string `json:"usuarioMode"`
	ID        string `json:"usuarioID"`
}

func usuarioEditAttributes(dialogID string, user *model.ResponseUsuario) templ.Attributes {
	roles := make([]map[string]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		if role != nil {
			roles = append(roles, map[string]string{"rol_id": role.RolID, "unidad_id": role.UnidadID, "rol_name": role.RolName, "unidad_name": role.UnidadName})
		}
	}
	rolesJSON, _ := json.Marshal(roles)
	return templ.Attributes{
		"tabindex":                   "0",
		"type":                       "button",
		"data-slot":                  "dialog-trigger",
		"data-base-ui-click-trigger": true,
		"data-templ-controls":        dialogID,
		"aria-haspopup":              "dialog",
		"aria-expanded":              "false",
		"data-on:click":              "" + usuarioEditScript(user, string(rolesJSON)),
	}
}

func usuarioEditScript(user *model.ResponseUsuario, rolesJSON string) string {
	js := func(value any) string {
		data, _ := json.Marshal(value)
		return string(data)
	}
	return "$usuarioMode = 'edit'; $usuarioID = " + js(user.ID) + "; $nombres = " + js(user.Nombres) + "; $apellido1 = " + js(user.Apellido1) + "; $apellido2 = " + js(optionalString(user.Apellido2)) + "; $documento = " + js(optionalString(user.Documento)) + "; $celular = " + js(optionalString(user.Celular)) + "; $correo = " + js(optionalString(user.Correo)) + "; $sexo = " + js(optionalString(user.Sexo)) + "; $direccion = " + js(optionalString(user.Direccion)) + "; $username = " + js(user.Username) + "; $password = ''; $foto64 = " + js("") + "; $rolesJson = " + js(rolesJSON) + "; const root = document.getElementById('usuario-dialog'); root?.querySelector('[data-usuario-title]')?.replaceChildren('Editar usuario'); root?.querySelector('#usuario-password')?.removeAttribute('required'); root?.querySelector('#usuario-submit')?.removeAttribute('disabled'); root?.querySelector('#usuario-submit')?.removeAttribute('aria-busy'); root?.querySelector('#usuario-submit')?.replaceChildren(document.createTextNode('Guardar usuario')); root?.querySelector('#usuario-form-error')?.replaceChildren(); const list = root?.querySelector('[data-usuario-roles-list]'); if (list) { list.innerHTML = ''; JSON.parse($rolesJson || '[]').forEach((pair) => { const row = document.createElement('div'); row.className = 'flex items-center justify-between gap-3 rounded-md border border-border/60 bg-background px-3 py-2 text-sm'; row.dataset.usuarioPair = pair.unidad_id; const text = document.createElement('span'); text.textContent = pair.rol_name + ' · ' + pair.unidad_name; const remove = document.createElement('button'); remove.type = 'button'; remove.className = 'text-xs text-muted-foreground hover:text-destructive'; remove.textContent = 'Quitar'; remove.onclick = () => { $rolesJson = JSON.stringify(JSON.parse($rolesJson || '[]').filter((item) => item.unidad_id !== row.dataset.usuarioPair)); row.remove(); }; row.append(text, remove); list.append(row); }); }"
}

func usuarioCreateResetScript() string {
	return "$usuarioMode = 'create'; $usuarioID = ''; $nombres = ''; $apellido1 = ''; $apellido2 = ''; $documento = ''; $celular = ''; $correo = ''; $sexo = ''; $direccion = ''; $username = ''; $password = ''; $foto64 = ''; $rolesJson = ''; const root = document.getElementById('usuario-dialog'); root?.querySelector('[data-usuario-title]')?.replaceChildren('Nuevo usuario'); root?.querySelector('#usuario-password')?.setAttribute('required', ''); root?.querySelector('#usuario-submit')?.removeAttribute('disabled'); root?.querySelector('#usuario-submit')?.removeAttribute('aria-busy'); root?.querySelector('#usuario-submit')?.replaceChildren(document.createTextNode('Guardar usuario')); root?.querySelector('[data-usuario-roles-list]')?.replaceChildren(); root?.querySelector('#usuario-form-error')?.replaceChildren(); root?.querySelector('[data-usuario-roles-error]')?.replaceChildren();"
}

func usuarioBind(signal string) templ.Attributes {
	return templ.Attributes{"data-bind:" + signal: true}
}

func usuarioInputType(signal string) string {
	if signal == "password" {
		return "password"
	}
	if signal == "correo" {
		return "email"
	}
	return "text"
}

func usuarioAddRoleUnitScript() string {
	return `const root = document.getElementById('usuario-dialog'); const error = root?.querySelector('[data-usuario-roles-error]'); if (!$rolForm || !$unidadForm) { if (error) error.innerHTML = '<p class="text-xs text-destructive">Seleccione un rol y una unidad.</p>'; return; }; const pairs = JSON.parse($rolesJson || '[]'); if (pairs.some((pair) => pair.unidad_id === $unidadForm)) { if (error) error.innerHTML = '<p class="text-xs text-destructive">La unidad ya tiene un rol asignado.</p>'; return; }; pairs.push({rol_id: $rolForm, unidad_id: $unidadForm}); $rolesJson = JSON.stringify(pairs); if (error) error.innerHTML = ''; const roleLabel = root?.querySelector('[data-usuario-role] option:checked')?.textContent || ''; const unitLabel = root?.querySelector('[data-usuario-unit] option:checked')?.textContent || ''; const list = root?.querySelector('[data-usuario-roles-list]'); if (!list) return; const row = document.createElement('div'); row.className = 'flex items-center justify-between gap-3 rounded-md border border-border/60 bg-background px-3 py-2 text-sm'; row.dataset.usuarioPair = $unidadForm; const text = document.createElement('span'); text.textContent = roleLabel + ' · ' + unitLabel; const remove = document.createElement('button'); remove.type = 'button'; remove.className = 'text-xs text-muted-foreground hover:text-destructive'; remove.textContent = 'Quitar'; remove.onclick = () => { const current = JSON.parse($rolesJson || '[]').filter((pair) => pair.unidad_id !== row.dataset.usuarioPair); $rolesJson = JSON.stringify(current); row.remove(); }; row.append(text, remove); list.append(row); $rolForm = ''; $unidadForm = '';`
}

func (c *Controller) UsuariosRender(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		c.usuarioCreate(w, r)
		return
	}
	data := c.usuariosPageData(r)
	if r.Header.Get("Datastar-Request") == "true" {
		c.usuariosPatch(w, r, data)
		return
	}

	_ = RenderUsuarios(data).Render(r.Context(), w)
}

func (c *Controller) usuariosPageData(r *http.Request) UsuariosPageData {
	page := parsePositiveInt(r.URL.Query().Get("pagina"), 1)
	size := parsePositiveInt(r.URL.Query().Get("size"), 10)
	rol := strings.TrimSpace(r.URL.Query().Get("rol"))

	data := UsuariosPageData{Pagina: page, Size: size, Rol: rol}
	query := model.QueryUsuarios{Pagina: page, Size: size}
	if rol != "" {
		query.Rol = &rol
	}
	data.Listado, _ = xrepo.Usuarios2(r.Context(), c.DB, query)
	if data.Listado == nil {
		data.Error = "No fue posible cargar los usuarios."
	}
	data.Roles, _ = xrepo.Roles(r.Context(), c.DB)
	data.Unidades, _ = xrepo.Unidades(r.Context(), c.DB)
	data.Detalles = usuariosDetalles(r, c.DB, data.Listado)
	return data
}

func usuariosDetalles(r *http.Request, db *sql.DB, listado *model.UsuarioPaginado) map[string]*model.ResponseUsuario {
	detalles := make(map[string]*model.ResponseUsuario)
	if listado == nil {
		return detalles
	}
	for _, user := range listado.Datos {
		if user == nil {
			continue
		}
		if detalle, err := xrepo.UsuarioByID(r.Context(), db, user.ID); err == nil && detalle != nil {
			detalles[user.ID] = detalle
		}
	}
	return detalles
}

func (c *Controller) usuarioCreate(w http.ResponseWriter, r *http.Request) {
	var signals usuarioFormSignals
	if err := datastar.ReadSignals(r, &signals); err != nil {
		_ = r.ParseForm()
		signals.Nombres = r.FormValue("nombres")
		signals.Apellido1 = r.FormValue("apellido1")
		signals.Apellido2 = r.FormValue("apellido2")
		signals.Documento = r.FormValue("documento")
		signals.Celular = r.FormValue("celular")
		signals.Correo = r.FormValue("correo")
		signals.Sexo = r.FormValue("sexo")
		signals.Direccion = r.FormValue("direccion")
		signals.Username = r.FormValue("username")
		signals.Password = r.FormValue("password")
		signals.RolesJSON = r.FormValue("rolesJson")
		signals.Mode = r.FormValue("usuarioMode")
		signals.ID = r.FormValue("usuarioID")
	}
	if signals.Mode == "" {
		signals.Mode = r.URL.Query().Get("mode")
	}
	if signals.ID == "" {
		signals.ID = r.URL.Query().Get("id")
	}

	roles := make([]*model.RolUnidad, 0)
	if signals.RolesJSON != "" {
		if err := json.Unmarshal([]byte(signals.RolesJSON), &roles); err != nil {
			roles = nil
		}
	}
	input := model.NewUsuario{
		Nombres: signals.Nombres, Apellido1: signals.Apellido1,
		Apellido2: usuarioOptional(signals.Apellido2), Documento: usuarioOptional(signals.Documento),
		Celular: usuarioOptional(signals.Celular), Correo: usuarioOptional(signals.Correo),
		Sexo: usuarioOptional(signals.Sexo), Direccion: usuarioOptional(signals.Direccion),
		Username: signals.Username, Password: signals.Password, Foto64: usuarioOptional(signals.Foto64),
		Roles: roles, PermisosSueltos: []string{}, MenusSueltos: []string{},
	}
	var err error
	if signals.Mode == "edit" && strings.TrimSpace(signals.ID) != "" {
		_, err = xrepo.UpdateUsuario(r.Context(), c.DB, model.UpdateUsuario{ID: signals.ID, Nombres: input.Nombres, Apellido1: input.Apellido1, Apellido2: input.Apellido2, Documento: input.Documento, Celular: input.Celular, Correo: input.Correo, Sexo: input.Sexo, Direccion: input.Direccion, Username: input.Username, Password: usuarioOptional(input.Password), Foto64: input.Foto64, Roles: input.Roles, PermisosSueltos: []string{}, MenusSueltos: []string{}})
	} else {
		_, err = xrepo.CreateUsuario(r.Context(), c.DB, input)
	}
	sse := datastar.NewSSE(w, r)
	if err != nil {
		var body strings.Builder
		_ = UsuarioFormError(err.Error()).Render(r.Context(), &body)
		_ = sse.PatchElements(body.String(), datastar.WithSelector("#usuario-form-error"), datastar.WithModeOuter())
		_ = sse.MarshalAndPatchSignals(map[string]bool{
			"usuarioSaving": false,
			"usuarioSaved":  false,
		})
		return
	}

	data := c.usuariosPageData(r)
	var table strings.Builder
	_ = UsuariosTableRegion(data).Render(r.Context(), &table)
	_ = sse.PatchElements(table.String(), datastar.WithSelector("#usuarios-table-region"), datastar.WithModeOuter())
	_ = sse.MarshalAndPatchSignals(map[string]bool{
		"usuarioSaving": false,
		"usuarioSaved":  true,
	})
}

func usuarioOptional(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
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

func usuarioInt(value int32) string {
	return strconv.Itoa(int(value))
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
		query := model.QueryUsuarios{Pagina: data.Pagina, Size: data.Size}
		if data.Rol != "" {
			query.Rol = &data.Rol
		}
		data.Listado, _ = xrepo.Usuarios2(r.Context(), c.DB, query)
		if data.Listado == nil {
			data.Error = "No fue posible cargar los usuarios."
		}
		data.Detalles = usuariosDetalles(r, c.DB, data.Listado)
	}

	var body strings.Builder
	_ = UsuariosTableRegion(data).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#usuarios-table-region"), datastar.WithModeOuter())
	_ = sse.ExecuteScript("history.replaceState({},''," + strconv.Quote(usuarioQueryURL(data)) + ")")
}

func usuarioQueryURL(data UsuariosPageData) string {
	query := "?pagina=" + strconv.Itoa(int(data.Pagina)) + "&size=" + strconv.Itoa(int(data.Size))
	if data.Rol != "" {
		query += "&rol=" + urlQueryEscape(data.Rol)
	}
	return ADMIN_USUARIOS + query
}

func urlQueryEscape(value string) string {
	return url.QueryEscape(value)
}

func parsePositiveInt(value string, fallback int32) int32 {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return int32(parsed)
}

func usuarioSignalString(value string) string {
	return strconv.Quote(value)
}

func usuarioSignalInt(value int32) string {
	return strconv.Itoa(int(value))
}

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
	return "$pagina = " + strconv.Itoa(int(page)) + "; @get('" + ADMIN_USUARIOS + "')"
}
