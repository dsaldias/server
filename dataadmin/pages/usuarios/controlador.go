package usuarios

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

type UsuariosController struct {
	DB *sql.DB
	C  *mainlayout.MainController
}

func (c *UsuariosController) Listar(w http.ResponseWriter, r *http.Request) {
	if _, err := utility.Is_Auth(c.DB, w, r, ""); err != nil {
		return
	}

	us, err := repo.Usuarios(r.Context(), c.DB, model.QueryUsuarios{})
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
		us, err := repo.UsuarioByID(r.Context(), c.DB, id)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		edit = *us
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

	Ver(edit, perms, mens).Render(r.Context(), w)
}

func (c *UsuariosController) FormNew(w http.ResponseWriter, r *http.Request) {
	if _, err := utility.Is_Auth(c.DB, w, r, ""); err != nil {
		return
	}
	id := chi.URLParam(r, "id")

	edit := model.ResponseUsuario{}

	if len(id) > 0 {
		us, err := repo.UsuarioByID(r.Context(), c.DB, id)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		edit = *us
	}

	rols, err := repo.Roles(r.Context(), c.DB)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	unis, err := repo.Unidades(r.Context(), c.DB)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
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

		_, err = repo.UpdateUsuario(r.Context(), c.DB, input)
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

		_, err = repo.CreateUsuario(r.Context(), c.DB, input)
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

func (c *UsuariosController) EditarPerfil(w http.ResponseWriter, r *http.Request) {
	xauth, err := utility.Is_Auth(c.DB, w, r, "")
	if err != nil {
		return
	}
	userid := xauth.Clains.USERID

	if r.Method == "GET" {
		us, err := repo.UsuarioByID(r.Context(), c.DB, userid)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		PerfilView(us).Render(r.Context(), w)

	} else {
		data, err := utility.ParseBodyToJSON(r)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}

		if data["foto64"] == nil || data["foto64"] == "" {
			data["foto64"] = nil
		}

		var input model.UpdatePerfil
		jsonData, err := json.Marshal(data)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}

		if err := json.Unmarshal(jsonData, &input); err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}

		_, err = repo.UpdatePerfil(r.Context(), c.DB, input)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
	}

}
