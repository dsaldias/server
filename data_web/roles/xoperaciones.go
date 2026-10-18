package roles

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

func (c *Controller) RolCrear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	c.rolSave(w, r, false)
}

func (c *Controller) RolEditar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	c.rolSave(w, r, true)
}

func (c *Controller) rolSave(w http.ResponseWriter, r *http.Request, editing bool) {
	var signals roleFormSignals
	if err := datastar.ReadSignals(r, &signals); err != nil {
		_ = r.ParseForm()
		signals.Nombre = r.FormValue("nombre")
		signals.Descripcion = r.FormValue("descripcion")
		signals.Jerarquia = parseRoleInt(r.FormValue("jerarquia"))
		signals.Permisos = parseRoleIDs(r.FormValue("permisos"))
		signals.Menus = parseRoleIDs(r.FormValue("menus"))
		signals.ID = r.FormValue("rolID")
	}
	input := model.NewRol{
		Nombre:      strings.TrimSpace(signals.Nombre),
		Descripcion: optionalString(signals.Descripcion),
		Jerarquia:   optionalInt(signals.Jerarquia),
		Permisos:    signals.Permisos,
		Menus:       roleMenuIDs(signals.Menus),
	}
	var err error
	if editing {
		if strings.TrimSpace(signals.ID) == "" {
			err = fmt.Errorf("El identificador del rol es obligatorio para editar.")
		} else {
			_, err = xrepo.UpdateRol(r.Context(), c.DB, model.UpdateRol{ID: signals.ID, Nombre: input.Nombre, Descripcion: input.Descripcion, Jerarquia: input.Jerarquia, Permisos: input.Permisos, Menus: input.Menus})
		}
	} else {
		_, err = xrepo.CreateRol(r.Context(), c.DB, input)
	}
	sse := datastar.NewSSE(w, r)
	if err != nil {
		_ = sse.PatchElements(roleFormError(err.Error()), datastar.WithSelector("#rol-form-error"), datastar.WithModeOuter())
		_ = sse.MarshalAndPatchSignals(map[string]bool{"rolSaving": false, "rolSaved": false, "rolesLoading": false})
		return
	}
	data := c.rolesTableData(r)
	var body strings.Builder
	_ = RolesTableRegion(data).Render(r.Context(), &body)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#roles-table-region"), datastar.WithModeOuter())
	_ = sse.MarshalAndPatchSignals(map[string]bool{"rolSaving": false, "rolSaved": true, "rolesLoading": false})
}

func parseRoleInt(value string) int32 {
	var result int32
	if _, err := fmt.Sscan(value, &result); err != nil {
		return 0
	}
	return result
}

func parseRoleIDs(value string) []string {
	var result []string
	if strings.TrimSpace(value) == "" {
		return result
	}
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return []string{}
	}
	return result
}

func roleMenuIDs(values []string) []int32 {
	menus := make([]int32, 0, len(values))
	for _, value := range values {
		var id int32
		if _, err := fmt.Sscan(value, &id); err == nil {
			menus = append(menus, id)
		}
	}
	return menus
}
