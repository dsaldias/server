package avisos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dsaldias/server/dataadmin/admin/mainlayout"
	"github.com/dsaldias/server/dataadmin/admin/utility"

	av "github.com/dsaldias/server/dataauth/avisos"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/go-chi/chi"
)

type AvisosController struct {
	DB *sql.DB
	C  *mainlayout.MainController
}

func (c *AvisosController) Listar(w http.ResponseWriter, r *http.Request) {
	if _, err := utility.Is_Auth(c.DB, w, r, ""); err != nil {
		return
	}

	us, err := av.GetNotificacionesActivas(c.DB)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	if r.URL.Query().Get("xrefresh") != "" {
		ListaAvisos(us).Render(r.Context(), w)
		return
	}

	contenido := Avisos(us, "?xrefresh=1")
	c.C.RenderPage(w, r, contenido)
}

func (c *AvisosController) FormNew(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	edit := model.Notificacion{}

	if len(id) > 0 {
		avi, err := av.Get(c.DB, id)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		edit = *avi
	}

	Formulario(edit).Render(r.Context(), w)
}

func (c *AvisosController) Crear(w http.ResponseWriter, r *http.Request) {
	xauth, err := utility.Is_Auth(c.DB, w, r, "")
	if err != nil {
		return
	}
	userid := xauth.Clains.USERID

	data, err := utility.ParseBodyToJSON(r)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	idavi := data["id"]

	data["desde"] = fmt.Sprintf("%s:00Z", data["desde"])
	data["hasta"] = fmt.Sprintf("%s:00Z", data["hasta"])

	jsonData, err := json.Marshal(data)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	if idavi != nil && idavi != "" {
		var input model.UpdNotificacion

		if err := json.Unmarshal(jsonData, &input); err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}

		_, err = av.Actualizar(c.DB, input, userid)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
	} else {
		var input model.NewNotificacion

		if err := json.Unmarshal(jsonData, &input); err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}

		_, err = av.Crear(c.DB, input, userid)
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

func (c *AvisosController) Ver(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	edit := model.Notificacion{}

	if len(id) > 0 {
		us, err := av.Get(c.DB, id)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		edit = *us
	}
	Ver(edit).Render(r.Context(), w)
}
