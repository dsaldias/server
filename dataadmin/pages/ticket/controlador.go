package ticket

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dsaldias/server/dataadmin/pages/mainlayout"
	"github.com/dsaldias/server/dataadmin/pages/utility"
	"github.com/dsaldias/server/dataauth/repo"

	"github.com/dsaldias/server/graph_auth/model"
	"github.com/go-chi/chi"
)

type TicketController struct {
	DB *sql.DB
	C  *mainlayout.MainController
}

func (c *TicketController) Listar(w http.ResponseWriter, r *http.Request) {
	us, err := repo.MisTickets(r.Context(), c.DB)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	q := model.QueryTickets{}
	all, err := repo.AllTickets(r.Context(), c.DB, q)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	rf := r.URL.Query().Get("xrefresh")
	if rf != "" {
		if rf == "2" {
			ListaTicketsAll(all).Render(r.Context(), w)
			return
		}
		ListaTickets(us).Render(r.Context(), w)
		return
	}

	contenido := Tickets(us, all, "?xrefresh=1", "?xrefresh=2")
	c.C.RenderPage(w, r, contenido)
}

func (c *TicketController) FormNew(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	edit := model.Ticket{}

	if len(id) > 0 {
		us, err := repo.VerTicket(r.Context(), c.DB, id)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		edit = *us
	}

	Formulario(edit).Render(r.Context(), w)
}

func (c *TicketController) Crear(w http.ResponseWriter, r *http.Request) {
	data, err := utility.ParseBodyToJSON(r)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	var input model.NewTicket

	if err := json.Unmarshal(jsonData, &input); err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	_, err = repo.CreateTicket(r.Context(), c.DB, input)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	q := r.URL.Query()
	q.Set("xrefresh", "1")
	r.URL.RawQuery = q.Encode()
	c.Listar(w, r)
}

func (c *TicketController) Ver(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	edit := model.Ticket{}

	if len(id) > 0 {
		us, err := repo.VerTicket(r.Context(), c.DB, id)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		edit = *us
	}
	Ver(edit).Render(r.Context(), w)
}

func (c *TicketController) FormResponder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	is_all := r.URL.Query().Get("is_from_all")

	edit := model.Ticket{}

	if len(id) > 0 {
		us, err := repo.VerTicket(r.Context(), c.DB, id)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		edit = *us
	} else {
		utility.ErrorResponse(w, r, errors.New("falta el id"), nil)
		return
	}

	Responder(edit, is_all == "1").Render(r.Context(), w)
}

func (c *TicketController) ResponderTicket(w http.ResponseWriter, r *http.Request) {
	data, err := utility.ParseBodyToJSON(r)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	var input model.NewTicketRespuesta

	if err := json.Unmarshal(jsonData, &input); err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	_, err = repo.UpdateTicket(r.Context(), c.DB, input)
	if err != nil {
		utility.ErrorResponse(w, r, err, nil)
		return
	}

	if data["is_from_all"] != nil {
		q := model.QueryTickets{}
		all, err := repo.AllTickets(r.Context(), c.DB, q)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		ListaTicketsAll(all).Render(r.Context(), w)

	} else {
		us, err := repo.MisTickets(r.Context(), c.DB)
		if err != nil {
			utility.ErrorResponse(w, r, err, nil)
			return
		}
		ListaTickets(us).Render(r.Context(), w)

	}
}
