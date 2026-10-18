package usuarios

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

func usuarioModalRequestScript(path string) string {
	return "const button = evt.currentTarget; if (button) { button.disabled = true; button.setAttribute('aria-busy','true'); } @get('" + path + "')"
}

func usuarioModalRequestWithIDScript(path string) string {
	return "const button = evt.currentTarget; if (button) { button.disabled = true; button.setAttribute('aria-busy','true'); } @get('" + path + "?userid=' + encodeURIComponent(button?.dataset.usuarioId || ''))"
}

func usuarioFormSubmitScript(user *model.ResponseUsuario) string {
	path := ADMIN_USUARIOS_CREAR
	if user != nil {
		path = ADMIN_USUARIOS_EDITAR
	}
	return "$usuarioSaving = true; @post('" + path + "?pagina='+$pagina+'&size='+$size+'&rol='+encodeURIComponent($rol || ''))"
}

func usuarioModalReadyScript() string {
	return "document.querySelectorAll('[data-usuario-modal-action]').forEach((button) => { button.disabled = false; button.removeAttribute('aria-busy'); });"
}

func usuarioDialogCleanupScript() string {
	return "const content = document.getElementById('usuario-dialog-content'); if (content) { content.className = 'flex min-h-40 items-center justify-center gap-2 text-sm text-muted-foreground'; content.innerHTML = '<span class=\"size-4 animate-spin rounded-full border-2 border-current border-t-transparent\"></span>Cargando, espere...'; } document.querySelectorAll('[data-usuario-modal-action]').forEach((button) => { button.disabled = false; button.removeAttribute('aria-busy'); }); const title = document.querySelector('#usuario-dialog [data-usuario-title]'); title?.replaceChildren(document.createTextNode('Usuario'));"
}

func usuarioModalTitleScript(title string) string {
	data, _ := json.Marshal(title)
	return "document.querySelector('#usuario-dialog [data-usuario-title]')?.replaceChildren(document.createTextNode(" + string(data) + "));"
}

func usuarioModalDescriptionScript(description string) string {
	data, _ := json.Marshal(description)
	return "document.querySelector('#usuario-dialog [data-templ-dialog-description]')?.replaceChildren(document.createTextNode(" + string(data) + "));"
}

func usuarioRequestID(r *http.Request) string {
	id := strings.TrimSpace(r.URL.Query().Get("userid"))
	if id == "" {
		id = strings.TrimSpace(r.FormValue("userid"))
	}
	return id
}

func usuarioFormMode(user *model.ResponseUsuario) string {
	if user == nil {
		return "create"
	}
	return "edit"
}

func usuarioFormID(user *model.ResponseUsuario) string {
	if user == nil {
		return ""
	}
	return user.ID
}

func usuarioFormValue(user *model.ResponseUsuario, field string) string {
	if user == nil {
		return ""
	}
	switch field {
	case "nombres":
		return user.Nombres
	case "apellido1":
		return user.Apellido1
	case "apellido2":
		return optionalStringValue(user.Apellido2)
	case "documento":
		return optionalStringValue(user.Documento)
	case "celular":
		return optionalStringValue(user.Celular)
	case "correo":
		return optionalStringValue(user.Correo)
	case "sexo":
		return optionalStringValue(user.Sexo)
	case "direccion":
		return optionalStringValue(user.Direccion)
	case "username":
		return user.Username
	default:
		return ""
	}
}

func usuarioFormRolesJSON(user *model.ResponseUsuario) string {
	if user == nil {
		return ""
	}
	roles, _ := json.Marshal(usuarioRolePairs(user))
	return string(roles)
}

func usuarioCreateResetScript() string {
	return "setTimeout(() => { $usuarioMode = 'create'; $usuarioID = ''; $nombres = ''; $apellido1 = ''; $apellido2 = ''; $documento = ''; $celular = ''; $correo = ''; $sexo = ''; $direccion = ''; $username = ''; $password = ''; $foto64 = ''; $rolesJson = ''; const root = document.getElementById('usuario-dialog'); root?.querySelector('[data-usuario-title]')?.replaceChildren('Nuevo usuario'); root?.querySelectorAll('input, textarea, select').forEach((field) => { field.value = ''; field.dispatchEvent(new Event('input', { bubbles: true })); }); root?.querySelector('#usuario-password')?.setAttribute('required', ''); root?.querySelector('#usuario-submit')?.removeAttribute('disabled'); root?.querySelector('#usuario-submit')?.removeAttribute('aria-busy'); root?.querySelector('#usuario-submit')?.replaceChildren(document.createTextNode('Guardar usuario')); root?.querySelector('[data-usuario-roles-list]')?.replaceChildren(); root?.querySelector('#usuario-form-error')?.replaceChildren(); root?.querySelector('[data-usuario-roles-error]')?.replaceChildren(); }, 0);"
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

func usuarioRemoveRoleUnitScript() string {
	return "const row = evt.currentTarget.closest('[data-usuario-pair]'); if (!row) return; $rolesJson = JSON.stringify(JSON.parse($rolesJson || '[]').filter((pair) => pair.unidad_id !== row.dataset.usuarioPair)); row.remove();"
}

func (c *Controller) UsuarioFormularioRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	roles, _ := xrepo.Roles(r.Context(), c.DB)
	unidades, _ := xrepo.Unidades(r.Context(), c.DB)
	c.patchUsuarioModal(w, r, Formulario(roles, unidades, nil), usuarioCreateResetScript()+usuarioModalReadyScript())
}

func (c *Controller) UsuarioVisualizarRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	detalle, err := xrepo.UsuarioByID(r.Context(), c.DB, usuarioRequestID(r))
	if err != nil || detalle == nil {
		c.patchUsuarioModal(w, r, UsuarioModalError("No fue posible cargar el usuario."), usuarioModalTitleScript("Usuario")+usuarioModalDescriptionScript("No fue posible cargar la información.")+usuarioModalReadyScript())
		return
	}
	c.patchUsuarioModal(w, r, Visualizar(detalle), usuarioModalTitleScript("Visualizar usuario")+usuarioModalDescriptionScript("Consulta la información registrada del usuario.")+usuarioModalReadyScript())
}

func (c *Controller) UsuarioEditarRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	detalle, err := xrepo.UsuarioByID(r.Context(), c.DB, usuarioRequestID(r))
	if err != nil || detalle == nil {
		c.patchUsuarioModal(w, r, UsuarioModalError("No fue posible cargar el usuario."), usuarioModalTitleScript("Usuario")+usuarioModalDescriptionScript("No fue posible cargar la información.")+usuarioModalReadyScript())
		return
	}
	roles, _ := xrepo.Roles(r.Context(), c.DB)
	unidades, _ := xrepo.Unidades(r.Context(), c.DB)
	c.patchUsuarioModal(w, r, Formulario(roles, unidades, detalle), usuarioModalTitleScript("Editar usuario")+usuarioModalDescriptionScript("Registra la información del usuario y asigna sus roles por unidad.")+usuarioModalReadyScript())
}

func (c *Controller) patchUsuarioModal(w http.ResponseWriter, r *http.Request, component templ.Component, after string) {
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
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#usuario-dialog-content"), datastar.WithModeOuter())
	if after != "" {
		_ = sse.ExecuteScript(after)
	}
}

func usuarioRolePairs(user *model.ResponseUsuario) []map[string]string {
	roles := make([]map[string]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		if role != nil {
			roles = append(roles, map[string]string{"rol_id": role.RolID, "unidad_id": role.UnidadID, "rol_name": role.RolName, "unidad_name": role.UnidadName})
		}
	}
	return roles
}
