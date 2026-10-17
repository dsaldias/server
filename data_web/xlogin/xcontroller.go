package xlogin

import (
	"bytes"
	"database/sql"
	"net/http"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type Controller struct {
	DB *sql.DB
}

func (c *Controller) LoginRender(w http.ResponseWriter, r *http.Request) {
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
		_ = sse.PatchElements(body.String(), datastar.WithSelector("#login-result"), datastar.WithModeOuter())
		return
	}

	var body bytes.Buffer
	InicioRender().Render(r.Context(), &body)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#page"), datastar.WithModeOuter())
}
