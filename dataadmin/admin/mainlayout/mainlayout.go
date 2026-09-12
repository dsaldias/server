package mainlayout

import (
	"database/sql"
	"net/http"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/dataadmin/admin/mainlayout/principal"
	"github.com/dsaldias/server/dataadmin/admin/utility"
	"github.com/dsaldias/server/dataauth/menus"
	"github.com/dsaldias/server/dataauth/roles"
	"github.com/dsaldias/server/dataauth/usuarios"
	"github.com/dsaldias/server/graph_auth/model"
)

type MainController struct {
	DB *sql.DB
}

func (c *MainController) SetCookieUnidadId(w http.ResponseWriter, r *http.Request) {

	unidad_id := r.FormValue("unidad_id")
	// rol_id := r.FormValue("rol_id")

	// fmt.Println("UNIDAD:", unidad_id)
	// fmt.Println("ROL:", rol_id)

	// AL PARECER EL ROL NO SE USA, POR MEDIO DE LA UNIDAD Y USERID YA SE VERIFICA EL PERMISO EN EL BACK
	utility.SetUnidadOnCookie(unidad_id, w, r)
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

	principal.MainPageLayout(
		"Proyecto academicos",
		unidadid,
		mens,
		rols,
		user,
		contenido,
		ruta,
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

	cookie, err := r.Cookie("galletita_traviesa_unidad_default")
	if err != nil {
		return
	}

	unidadid = cookie.Value

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
