package usuarios

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/dsaldias/server/dataadmin/admin/mainlayout"
	"github.com/dsaldias/server/dataadmin/admin/utility"

	"github.com/dsaldias/server/dataauth/menus"
	"github.com/dsaldias/server/dataauth/permisos"
	"github.com/dsaldias/server/dataauth/roles"
	"github.com/dsaldias/server/dataauth/unidades"
	"github.com/dsaldias/server/dataauth/usuarios"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/go-chi/chi"
)

type UsuariosController struct {
	DB *sql.DB
	C  *mainlayout.MainController
}

func (c *UsuariosController) Listar(w http.ResponseWriter, r *http.Request) {
	if _, err := utility.Is_Auth(c.DB, w, r, ""); err != nil {
		return
	}

	us, err := usuarios.GetUsuarios(c.DB, model.QueryUsuarios{})
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	if r.URL.Query().Get("xrefresh") != "" {
		ListaUsuarios(us).Render(r.Context(), w)
		return
	}

	contenido := Usuarios(us, "?xrefresh=1")
	c.C.RenderPage(w, r, contenido)
}

func (c *UsuariosController) Ver(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	edit := model.ResponseUsuario{}

	if len(id) > 0 {
		us, err := usuarios.GetBy(c.DB, id)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		edit = *us
	}

	perms, err := permisos.GetPermisos(c.DB)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	mens, err := menus.Listar(c.DB)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	Ver(edit, perms, mens).Render(r.Context(), w)
}

func (c *UsuariosController) FormNew(w http.ResponseWriter, r *http.Request) {
	if _, err := utility.Is_Auth(c.DB, w, r, ""); err != nil {
		return
	}
	id := chi.URLParam(r, "id")

	edit := model.ResponseUsuario{}

	if len(id) > 0 {
		us, err := usuarios.GetBy(c.DB, id)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		edit = *us
	}

	rols, err := roles.GetRoles(c.DB)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	unis, err := unidades.Listar(c.DB)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	perms, err := permisos.GetPermisos(c.DB)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	mens, err := menus.Listar(c.DB)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	Formulario(edit, rols, unis, perms, mens).Render(r.Context(), w)
}

func (c *UsuariosController) Crear(w http.ResponseWriter, r *http.Request) {
	data, err := utility.ParseBodyToJSON(r)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	iduser := data["id"]

	jsonData, err := json.Marshal(data)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	if iduser != nil && iduser != "" {
		var input model.UpdateUsuario

		if err := json.Unmarshal(jsonData, &input); err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}

		_, err = usuarios.Actualizar(c.DB, input)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
	} else {
		var input model.NewUsuario

		if err := json.Unmarshal(jsonData, &input); err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}

		_, err = usuarios.Crear(c.DB, input, nil)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
	}

	q := r.URL.Query()
	q.Set("xrefresh", "1")
	r.URL.RawQuery = q.Encode()
	c.Listar(w, r)
}
