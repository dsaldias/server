package xlogin

import (
	"bytes"
	"database/sql"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/data_web/components/combobox"
	"github.com/dsaldias/server/dataauth/utilidad"
	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type Controller struct {
	DB *sql.DB
}

type InicioData struct {
	Authenticated bool
	Me            *model.ResponseMe
	UnidadID      string
	RolID         string
}

func (c *Controller) LoginRender(w http.ResponseWriter, r *http.Request) {
	if c.authenticated(r) {
		http.Redirect(w, r, ADMIN_INICIO, http.StatusSeeOther)
		return
	}

	ctx := templ.WithChildren(r.Context(), LoginRender())
	MainLayout("Acceso | Panel administrativo", Default()).Render(ctx, w)
}

func (c *Controller) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	var input model.NewLogin2
	if err := datastar.ReadSignals(r, &input); err != nil {
		_ = r.ParseForm()
		input.Username = r.FormValue("username")
		input.Password = r.FormValue("password")
	}

	result, err := xrepo.LoginV2(r.Context(), c.DB, input)
	sse := datastar.NewSSE(w, r)
	if err != nil || result == nil {
		message := "No fue posible iniciar sesión."
		if err != nil {
			message = err.Error()
		}
		var body bytes.Buffer
		LoginError(message).Render(r.Context(), &body)
		selector := "#login-result"
		if r.URL.Query().Get("mode") == "relogin" {
			selector = "#relogin-result"
		}
		_ = sse.PatchElements(body.String(), datastar.WithSelector(selector), datastar.WithModeOuter())
		return
	}

	if r.URL.Query().Get("mode") == "relogin" {
		_ = sse.ExecuteScript("document.getElementById('relogin-dialog')?.remove()")
		return
	}

	_ = sse.Redirect(ADMIN_INICIO)
}

func (c *Controller) Logout(w http.ResponseWriter, r *http.Request) {
	cookies := []string{
		"galletita_traviesa",
	}

	for _, name := range cookies {
		http.SetCookie(w, &http.Cookie{
			Name:   name,
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
	}

	http.Redirect(w, r, ADMIN_HOME, http.StatusSeeOther)
}

func (c *Controller) InicioRender(w http.ResponseWriter, r *http.Request) {
	ctx := templ.WithChildren(r.Context(), InicioRender(c.inicioData(r)))
	MainLayout("Acceso | Panel administrativo", Default()).Render(ctx, w)
}

func (c *Controller) inicioData(r *http.Request) InicioData {
	auth, err := utilidad.CtxValue(r.Context(), c.DB, "")
	if err != nil {
		return InicioData{}
	}

	data := InicioData{
		Authenticated: true,
		UnidadID:      auth.UnidadID,
		RolID:         auth.RolID,
	}
	data.Me, _ = xrepo.Me(r.Context(), c.DB, model.InputMe{UnidadID: auth.UnidadID})
	return data
}

func (c *Controller) authenticated(r *http.Request) bool {
	_, err := utilidad.CtxValue(r.Context(), c.DB, "")
	return err == nil
}

func inicioUserName(me *model.ResponseMe) string {
	if me == nil || me.Usuario == nil {
		return "Usuario"
	}
	return strings.TrimSpace(me.Usuario.Nombres + " " + me.Usuario.Apellido1)
}

func inicioUserInitials(me *model.ResponseMe) string {
	name := inicioUserName(me)
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "U"
	}
	if len(parts) == 1 {
		return strings.ToUpper(string([]rune(parts[0])[0]))
	}
	return strings.ToUpper(string([]rune(parts[0])[0]) + string([]rune(parts[len(parts)-1])[0]))
}

func inicioSelectedRole(me *model.ResponseMe, unidadID, rolID string) string {
	if me != nil {
		for _, role := range me.Roles {
			if role.Unidad != nil && role.Rol != nil && role.Unidad.ID == unidadID && role.Rol.ID == rolID {
				return role.Unidad.Nombre + " / " + role.Rol.Nombre
			}
		}
	}
	return "Seleccionar unidad / rol"
}

func inicioRoleItems(me *model.ResponseMe) []combobox.ItemData {
	if me == nil {
		return nil
	}

	items := make([]combobox.ItemData, 0, len(me.Roles))
	for _, role := range me.Roles {
		if role.Unidad == nil || role.Rol == nil {
			continue
		}
		items = append(items, combobox.ItemData{
			Value: role.Unidad.ID + "|" + role.Rol.ID,
			Label: role.Unidad.Nombre + " / " + role.Rol.Nombre,
		})
	}
	return items
}
