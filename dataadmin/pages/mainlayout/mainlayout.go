package mainlayout

import (
	"database/sql"
	"net/http"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/dataadmin/pages/mainlayout/layoutconfig"
	"github.com/dsaldias/server/dataadmin/pages/mainlayout/principal"
	"github.com/dsaldias/server/dataadmin/pages/utility"
	"github.com/dsaldias/server/dataauth/menus"
	"github.com/dsaldias/server/dataauth/repo"
	"github.com/dsaldias/server/dataauth/roles"
	"github.com/dsaldias/server/dataauth/usuarios"
	"github.com/dsaldias/server/dataauth/utils"
	"github.com/dsaldias/server/graph_auth/model"
)

type MainController struct {
	DB     *sql.DB
	Config layoutconfig.Config
}

func (c *MainController) SetCookieUnidadId(w http.ResponseWriter, r *http.Request) {

	rol_id := r.FormValue("rol_id")
	unidad_id := r.FormValue("unidad_id")

	cookie, err := utils.CtxGetCookie(r)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)

	} else {
		cookie.RolID = rol_id
		cookie.UnidadID = unidad_id
		utils.CtxSetCookie(r.Context(), *cookie)

		mens, err := repo.GetMenusbyRol(r.Context(), c.DB, cookie.RolID)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		principal.MenuItems(mens, nil, "").Render(r.Context(), w)
	}

}

func (c *MainController) MainLayout(w http.ResponseWriter, r *http.Request) {
	c.RenderLayout(w, r, principal.Inicio())
}

func (c *MainController) RenderLayout(
	w http.ResponseWriter,
	r *http.Request,
	contenido templ.Component,
) {
	_, unidadid, mens, rols, user, err := c.layoutData(w, r)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	ruta := r.URL.Path

	config := c.Config.WithDefaults()
	title := config.Title

	principal.MainPageLayout(
		title,
		unidadid,
		mens,
		rols,
		user,
		contenido,
		ruta,
		config,
	).Render(r.Context(), w)
}

func (c *MainController) RenderPage(
	w http.ResponseWriter,
	r *http.Request,
	contenido templ.Component,
) {
	if r.Header.Get("HX-Request") == "true" {
		contenido.Render(r.Context(), w)
		return
	}

	c.RenderLayout(w, r, contenido)
}

func (c *MainController) layoutData(
	w http.ResponseWriter,
	r *http.Request,
) (
	userid string,
	unidadid string,
	mens []*model.Menus,
	rols []*model.ResponseRolMe,
	user *model.ResponseUsuario,
	err error,
) {
	xauth, err := utility.Is_Auth(c.DB, w, r, "")
	if err != nil {
		return
	}
	userid = xauth.Clains.USERID

	cookie, err := utils.CtxGetCookie(r)
	if err != nil {
		return
	}

	unidadid = cookie.UnidadID

	mens, err = menus.GetMenusbyRol(c.DB, unidadid)
	if err != nil {
		return
	}

	rols, err = roles.GetRolesByUsuario(c.DB, userid)
	if err != nil {
		return
	}

	user, err = usuarios.GetBy(c.DB, userid)
	return
}
