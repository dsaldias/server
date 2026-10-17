package xlogin

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
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
	Content       templ.Component
}

type contextInput struct {
	Context string `json:"context"`
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

func (c *Controller) renderModuleRoute(w http.ResponseWriter, r *http.Request, module http.Handler) {
	if r.Header.Get("X-Requested-With") == "fetch" {
		module.ServeHTTP(w, r)
		return
	}

	recorder := httptest.NewRecorder()
	module.ServeHTTP(recorder, r)
	data := c.inicioData(r)
	content := append([]byte(nil), recorder.Body.Bytes()...)
	data.Content = templ.ComponentFunc(func(_ context.Context, target io.Writer) error {
		_, err := target.Write(content)
		return err
	})

	if recorder.Code >= http.StatusBadRequest {
		w.WriteHeader(recorder.Code)
	}
	ctx := templ.WithChildren(r.Context(), InicioRender(data))
	_ = MainLayout("Panel administrativo", Default()).Render(ctx, w)
}

func (c *Controller) ContextSubmit(w http.ResponseWriter, r *http.Request) {
	var input contextInput
	if err := datastar.ReadSignals(r, &input); err != nil {
		_ = r.ParseForm()
		input.Context = r.FormValue("context")
	}

	parts := strings.SplitN(input.Context, "|", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		c.contextError(w, r, "La unidad y el rol seleccionados no son válidos.")
		return
	}

	unidadID, rolID := parts[0], parts[1]
	me, err := xrepo.Me(r.Context(), c.DB, model.InputMe{UnidadID: unidadID})
	if err != nil || me == nil || !inicioHasRole(me, unidadID, rolID) {
		c.contextError(w, r, "No tienes acceso a la unidad y rol seleccionados.")
		return
	}

	menus, err := xrepo.GetMenusbyRol(r.Context(), c.DB, rolID)
	if err != nil {
		c.contextError(w, r, "No fue posible cargar los menús del rol seleccionado.")
		return
	}

	cookie, err := utilidad.CtxGetCookie(r)
	if err != nil {
		c.contextError(w, r, "La sesión ya no está disponible. Inicia sesión nuevamente.")
		return
	}
	cookie.RolID = rolID
	cookie.UnidadID = unidadID
	utilidad.CtxSetCookie(r.Context(), *cookie)

	data := InicioData{Authenticated: true, Me: me, UnidadID: unidadID, RolID: rolID}
	data.Me.Menus = menus

	var body bytes.Buffer
	if err := InicioSidebar(data).Render(r.Context(), &body); err != nil {
		c.contextError(w, r, "No fue posible actualizar el menú lateral.")
		return
	}

	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#admin-sidebar"), datastar.WithModeOuter())
	_ = sse.ExecuteScript("const role = document.getElementById('active-role'); if (role) role.textContent = " + strconv.Quote(inicioSelectedRole(me, unidadID, rolID)) + "; const result = document.getElementById('context-result'); if (result) result.replaceChildren()")
}

func (c *Controller) contextError(w http.ResponseWriter, r *http.Request, message string) {
	var body bytes.Buffer
	_ = InicioContextError(message).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#context-result"), datastar.WithModeOuter())
}

func inicioHasRole(me *model.ResponseMe, unidadID, rolID string) bool {
	if me == nil {
		return false
	}
	for _, role := range me.Roles {
		if role != nil && role.Unidad != nil && role.Rol != nil && role.Unidad.ID == unidadID && role.Rol.ID == rolID {
			return true
		}
	}
	return false
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
	if data.Me != nil {
		// La cookie define el rol activo. Reemplazamos los menús generales
		// devueltos por Me por los que pertenecen a ese rol concreto.
		if menus, menuErr := xrepo.GetMenusbyRol(r.Context(), c.DB, auth.RolID); menuErr == nil {
			data.Me.Menus = menus
		}
	}
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

func inicioRootMenus(menus []*model.Menus) []*model.Menus {
	roots := make([]*model.Menus, 0)
	for _, menu := range menus {
		if menu != nil && menu.PadreID == nil {
			roots = append(roots, menu)
		}
	}
	return roots
}

func inicioChildMenus(menus []*model.Menus, parentID string) []*model.Menus {
	children := make([]*model.Menus, 0)
	for _, menu := range menus {
		if menu != nil && menu.PadreID != nil && *menu.PadreID == parentID {
			children = append(children, menu)
		}
	}
	return children
}

func inicioMenuClick(menu *model.Menus) string {
	if menu == nil {
		return ""
	}

	path := strconv.Quote(menu.Path)
	return fmt.Sprintf("(async()=>{const target=document.getElementById('admin-content');if(!target)return;const path=%s;target.innerHTML='<div class=\\\"flex min-h-48 items-center justify-center text-sm text-muted-foreground\\\">Cargando...</div>';try{const response=await fetch(path,{headers:{'Accept':'text/html','X-Requested-With':'fetch'}});const html=await response.text();const parsed=new DOMParser().parseFromString(html,'text/html');const page=parsed.querySelector('#admin-content');target.innerHTML=page?page.innerHTML:html;if(response.ok)window.history.pushState({},'',path)}catch(error){target.innerHTML='<div class=\\\"rounded-lg border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive\\\">No fue posible cargar el contenido: '+error.message+'</div>'}})();", path)
}
