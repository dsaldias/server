package unidades

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

type UnidadesController struct {
	DB *sql.DB
	C  *mainlayout.MainController
}

func (c *UnidadesController) Listar(w http.ResponseWriter, r *http.Request) {
	if _, err := utility.Is_Auth(c.DB, w, r, ""); err != nil {
		return
	}

	pagina, size := utility.GetPaginacionParams(r)

	q := model.QueryUnidades{
		Pagina: pagina,
		Size:   size,
	}
	unpag, err := repo.Unidades2(r.Context(), c.DB, q)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	pag := utility.GetPaginacion(r, unpag.Paginacion.Paginas)
	pag.TargetHtmx = "#lista-unidades"

	if utility.IsOnlyHtmx(r) {
		TablaUnidades(unpag.Datos, pag).Render(r.Context(), w)
		return
	}

	contenido := Unidades(unpag.Datos, pag, "?xrefresh=1")
	c.C.RenderPage(w, r, contenido)
}

func (c *UnidadesController) FormNew(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	edit := model.Unidad{}

	if len(id) > 0 {
		us, err := repo.UnidadByID(r.Context(), c.DB, id)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		edit = *us
	}

	Formulario(edit).Render(r.Context(), w)
}

func (c *UnidadesController) Crear(w http.ResponseWriter, r *http.Request) {
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
		var input model.UpdUnidad

		if err := json.Unmarshal(jsonData, &input); err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}

		_, err = repo.UpdateUnidad(r.Context(), c.DB, input)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
	} else {
		var input model.NewUnidad

		if err := json.Unmarshal(jsonData, &input); err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}

		_, err = repo.CreateUnidad(r.Context(), c.DB, input)
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

func (c *UnidadesController) Ver(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	edit := model.Unidad{}

	if len(id) > 0 {
		us, err := repo.UnidadByID(r.Context(), c.DB, id)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		edit = *us
	}
	Ver(edit).Render(r.Context(), w)
}
