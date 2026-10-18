package usuarios

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

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

func (c *Controller) UsuarioCrear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	c.usuarioSave(w, r, false)
}

func (c *Controller) UsuarioEditar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	c.usuarioSave(w, r, true)
}

func (c *Controller) usuarioSave(w http.ResponseWriter, r *http.Request, editing bool) {
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
		signals.Foto64 = r.FormValue("foto64")
		signals.RolesJSON = r.FormValue("rolesJson")
		signals.Mode = r.FormValue("usuarioMode")
		signals.ID = r.FormValue("usuarioID")
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
	if editing {
		if strings.TrimSpace(signals.ID) == "" {
			err = fmt.Errorf("El identificador del usuario es obligatorio para editar.")
		} else {
			_, err = xrepo.UpdateUsuario(r.Context(), c.DB, model.UpdateUsuario{ID: signals.ID, Nombres: input.Nombres, Apellido1: input.Apellido1, Apellido2: input.Apellido2, Documento: input.Documento, Celular: input.Celular, Correo: input.Correo, Sexo: input.Sexo, Direccion: input.Direccion, Username: input.Username, Password: usuarioOptional(input.Password), Foto64: input.Foto64, Roles: input.Roles, PermisosSueltos: []string{}, MenusSueltos: []string{}})
		}
	} else {
		_, err = xrepo.CreateUsuario(r.Context(), c.DB, input)
	}
	sse := datastar.NewSSE(w, r)
	if err != nil {
		var body strings.Builder
		_ = UsuarioFormError(err.Error()).Render(r.Context(), &body)
		_ = sse.PatchElements(body.String(), datastar.WithSelector("#usuario-form-error"), datastar.WithModeOuter())
		_ = sse.MarshalAndPatchSignals(map[string]bool{"usuarioSaving": false, "usuarioSaved": false})
		return
	}

	data := c.usuariosTableData(r)
	var table strings.Builder
	_ = UsuariosTableRegion(data).Render(r.Context(), &table)
	_ = sse.PatchElements(table.String(), datastar.WithSelector("#usuarios-table-region"), datastar.WithModeOuter())
	_ = sse.MarshalAndPatchSignals(map[string]bool{"usuarioSaving": false, "usuarioSaved": true})
}

func usuarioOptional(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
