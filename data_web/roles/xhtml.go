package roles

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type roleFormSignals struct {
	Nombre      string   `json:"nombre"`
	Descripcion string   `json:"descripcion"`
	Jerarquia   int32    `json:"jerarquia"`
	Permisos    []string `json:"permisos"`
	Menus       []string `json:"menus"`
	ID          string   `json:"rolID"`
}

func roleModalRequestScript(path string) string {
	return "const button = evt.currentTarget; if (button) { button.disabled = true; button.setAttribute('aria-busy','true'); } @get('" + path + "')"
}

func roleModalRequestWithIDScript(path string) string {
	return "const button = evt.currentTarget; const roleID = button?.getAttribute('data-role-id') || ''; if (button) { button.disabled = true; button.setAttribute('aria-busy','true'); } @get('" + path + "?roleid=' + encodeURIComponent(roleID) + '&id=' + encodeURIComponent(roleID))"
}

func roleFormSubmitScript(role *model.Rol) string {
	path := ADMIN_ROLES_CREAR
	if role != nil {
		path = ADMIN_ROLES_EDITAR
	}
	return "$rolSaving = true; @post('" + path + "?pagina='+$pagina+'&size='+$size)"
}

func roleModalLoading() string {
	return "const content = document.getElementById('rol-dialog-content'); if (content) { content.className = 'flex min-h-40 items-center justify-center gap-2 text-sm text-muted-foreground'; content.innerHTML = '<span class=\"size-4 animate-spin rounded-full border-2 border-current border-t-transparent\"></span>Cargando, espere...'; } document.querySelectorAll('[data-rol-modal-action]').forEach((button) => { button.disabled = false; button.removeAttribute('aria-busy'); });"
}

func roleModalReadyScript() string {
	return "document.querySelectorAll('[data-rol-modal-action]').forEach((button) => { button.disabled = false; button.removeAttribute('aria-busy'); });"
}

func roleModalTitleScript(title string) string {
	data, _ := json.Marshal(title)
	return "document.querySelector('#rol-dialog [data-rol-title]')?.replaceChildren(document.createTextNode(" + string(data) + "));"
}

func roleModalDescriptionScript(description string) string {
	data, _ := json.Marshal(description)
	return "document.querySelector('#rol-dialog [data-templ-dialog-description]')?.replaceChildren(document.createTextNode(" + string(data) + "));"
}

func roleRequestID(r *http.Request) string {
	id := strings.TrimSpace(r.URL.Query().Get("roleid"))
	if id == "" {
		id = strings.TrimSpace(r.URL.Query().Get("id"))
	}
	if id == "" {
		id = strings.TrimSpace(r.FormValue("roleid"))
	}
	if id == "" {
		id = strings.TrimSpace(r.FormValue("id"))
	}
	return id
}

func roleFormMode(role *model.Rol) string {
	if role == nil {
		return "create"
	}
	return "edit"
}

func roleFormID(role *model.Rol) string {
	if role == nil {
		return ""
	}
	return role.ID
}

func roleFormValue(role *model.Rol, field string) string {
	if role == nil {
		return ""
	}
	switch field {
	case "nombre":
		return role.Nombre
	case "descripcion":
		return optionalStringValue(role.Descripcion)
	default:
		return ""
	}
}

func roleFormInt(role *model.Rol) string {
	if role == nil {
		return "0"
	}
	return roleSignalInt(role.Jerarquia)
}

func roleFormIDs(role *model.Rol, menus bool) string {
	data, _ := json.Marshal(roleFormSelectionIDs(role, menus))
	return string(data)
}

func roleFormSelectionIDs(role *model.Rol, menus bool) []string {
	values := make([]string, 0)
	if role != nil {
		if menus {
			for _, item := range role.Menus {
				if item != nil {
					values = append(values, item.ID)
				}
			}
		} else {
			for _, item := range role.Permisos {
				if item != nil {
					values = append(values, item.Metodo)
				}
			}
		}
	}
	return values
}

func (c *Controller) RolFormularioRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	permisos, _ := xrepo.Permisos(r.Context(), c.DB)
	menus, _ := xrepo.Menus(r.Context(), c.DB)
	c.patchRoleModal(w, r, RoleForm(permisos, menus, nil), roleModalTitleScript("Nuevo rol")+roleModalDescriptionScript("Define la información, permisos y menús asignados al rol.")+roleModalReadyScript())
}

func (c *Controller) RolVisualizarRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	roleID := roleRequestID(r)
	if roleID == "" {
		c.patchRoleModal(w, r, RoleModalError("No se recibió el identificador del rol."), roleModalTitleScript("Rol")+roleModalDescriptionScript("Falta el identificador para consultar la información.")+roleModalReadyScript())
		return
	}
	role, err := xrepo.RolByID(r.Context(), c.DB, roleID)
	if err != nil || role == nil {
		c.patchRoleModal(w, r, RoleModalError("No fue posible cargar el rol."), roleModalTitleScript("Rol")+roleModalDescriptionScript("No fue posible cargar la información.")+roleModalReadyScript())
		return
	}
	c.patchRoleModal(w, r, RoleView(role), roleModalTitleScript("Visualizar rol")+roleModalDescriptionScript("Consulta la información y asignaciones del rol.")+roleModalReadyScript())
}

func (c *Controller) RolEditarRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	roleID := roleRequestID(r)
	if roleID == "" {
		c.patchRoleModal(w, r, RoleModalError("No se recibió el identificador del rol."), roleModalTitleScript("Editar rol")+roleModalDescriptionScript("Falta el identificador para cargar la información.")+roleModalReadyScript())
		return
	}
	role, err := xrepo.RolByID(r.Context(), c.DB, roleID)
	if err != nil || role == nil {
		c.patchRoleModal(w, r, RoleModalError("No fue posible cargar el rol."), roleModalTitleScript("Rol")+roleModalDescriptionScript("No fue posible cargar la información.")+roleModalReadyScript())
		return
	}
	permisos, _ := xrepo.Permisos(r.Context(), c.DB)
	menus, _ := xrepo.Menus(r.Context(), c.DB)
	after := roleModalTitleScript("Editar rol") + roleModalDescriptionScript("Define la información, permisos y menús asignados al rol.") + roleModalReadyScript()
	c.patchRoleModal(w, r, RoleForm(permisos, menus, role), after)
}

func (c *Controller) patchRoleModal(w http.ResponseWriter, r *http.Request, component templ.Component, after string) {
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
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#rol-dialog-content"), datastar.WithModeOuter())
	if after != "" {
		_ = sse.ExecuteScript(after)
	}
}
