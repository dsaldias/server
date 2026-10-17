package xlogin

import (
	"bytes"
	"database/sql"
	"net/http"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/dataauth/utilidad"
	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type Controller struct {
	DB *sql.DB
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

func (c *Controller) InicioRender(w http.ResponseWriter, r *http.Request) {
	ctx := templ.WithChildren(r.Context(), InicioRender(c.authenticated(r)))
	MainLayout("Acceso | Panel administrativo", Default()).Render(ctx, w)
}

func (c *Controller) authenticated(r *http.Request) bool {
	_, err := utilidad.CtxValue(r.Context(), c.DB, "")
	return err == nil
}
