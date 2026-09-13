package roles

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/dsaldias/server/dataadmin/pages/mainlayout"
	"github.com/dsaldias/server/dataadmin/pages/utility"
	"github.com/dsaldias/server/dataauth/repo"

	"github.com/dsaldias/server/graph_auth/model"
	"github.com/go-chi/chi"
)

type RolesController struct {
	DB *sql.DB
	C  *mainlayout.MainController
}

func (c *RolesController) ListarRoles(w http.ResponseWriter, r *http.Request) {
	if _, err := utility.Is_Auth(c.DB, w, r, ""); err != nil {
		return
	}

	roles, err := repo.Roles(r.Context(), c.DB)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	if r.URL.Query().Get("xrefresh") != "" {
		ListaRoles(roles).Render(r.Context(), w)
		return
	}

	contenido := Roles(roles, "?xrefresh=1")
	c.C.RenderPage(w, r, contenido)

}

func (c *RolesController) FormNew(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	edit := model.Rol{}

	if len(id) > 0 {
		rol, err := repo.RolByID(r.Context(), c.DB, id)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		edit = *rol
	}

	perms, err := repo.Permisos(r.Context(), c.DB)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	mens, err := repo.Menus(r.Context(), c.DB)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	Formulario(edit, mens, perms).Render(r.Context(), w)
}

func (c *RolesController) Crear(w http.ResponseWriter, r *http.Request) {
	data, err := utility.ParseBodyToJSON(r)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	idrol := data["id"]

	jsonData, err := json.Marshal(data)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	if idrol != nil && idrol != "" {
		var input model.UpdateRol

		if err := json.Unmarshal(jsonData, &input); err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}

		_, err = repo.UpdateRol(r.Context(), c.DB, input)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
	} else {
		var input model.NewRol

		if err := json.Unmarshal(jsonData, &input); err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}

		_, err = repo.CreateRol(r.Context(), c.DB, input)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
	}

	q := r.URL.Query()
	q.Set("xrefresh", "1")
	r.URL.RawQuery = q.Encode()
	c.ListarRoles(w, r)
}

func (c *RolesController) Ver(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	edit := model.Rol{}

	if len(id) > 0 {
		us, err := repo.RolByID(r.Context(), c.DB, id)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		edit = *us
	}

	Ver(edit).Render(r.Context(), w)
}
