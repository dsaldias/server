package tickets

import (
	"net/http"
	"strings"

	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type ticketCreateSignals struct {
	Problema string `json:"ticketProblema"`
}

type ticketReplySignals struct {
	ID        string `json:"ticketID"`
	Respuesta string `json:"ticketRespuesta"`
}

func (c *Controller) TicketCrear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	var signals ticketCreateSignals
	if err := datastar.ReadSignals(r, &signals); err != nil {
		_ = r.ParseForm()
		signals.Problema = r.FormValue("problema")
	}
	signals.Problema = strings.TrimSpace(signals.Problema)
	if signals.Problema == "" {
		c.ticketOperationError(w, r, "ticket-form-error", "La descripción del problema es obligatoria.")
		return
	}
	if _, err := xrepo.CreateTicket(r.Context(), c.DB, model.NewTicket{Problema: signals.Problema}); err != nil {
		c.ticketOperationError(w, r, "ticket-form-error", err.Error())
		return
	}
	c.ticketOperationSuccess(w, r, true)
}

func (c *Controller) TicketResponder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	var signals ticketReplySignals
	if err := datastar.ReadSignals(r, &signals); err != nil {
		_ = r.ParseForm()
		signals.ID = r.FormValue("ticketID")
		signals.Respuesta = r.FormValue("respuesta")
	}
	if strings.TrimSpace(signals.ID) == "" {
		signals.ID = strings.TrimSpace(r.URL.Query().Get("ticketid"))
	}
	signals.ID = strings.TrimSpace(signals.ID)
	signals.Respuesta = strings.TrimSpace(signals.Respuesta)
	if signals.ID == "" || signals.Respuesta == "" {
		c.ticketOperationError(w, r, "ticket-response-form-error", "La respuesta es obligatoria.")
		return
	}
	if _, err := xrepo.UpdateTicket(r.Context(), c.DB, model.NewTicketRespuesta{TicketsID: signals.ID, Respuesta: signals.Respuesta}); err != nil {
		c.ticketOperationError(w, r, "ticket-response-form-error", err.Error())
		return
	}
	c.ticketOperationSuccess(w, r, true)
}

func (c *Controller) TicketCerrar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("ticketid"))
	if id == "" {
		var signals ticketReplySignals
		if err := datastar.ReadSignals(r, &signals); err == nil {
			id = strings.TrimSpace(signals.ID)
		}
	}
	if id == "" {
		c.ticketOperationError(w, r, "ticket-response-form-error", "No se recibió el identificador del ticket.")
		return
	}
	if _, err := xrepo.CerrarTicket(r.Context(), c.DB, id); err != nil {
		c.ticketOperationError(w, r, "ticket-response-form-error", err.Error())
		return
	}
	c.ticketOperationSuccess(w, r, true)
}

func (c *Controller) ticketOperationError(w http.ResponseWriter, r *http.Request, errorID, message string) {
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(ticketFormErrorWithID(errorID, message), datastar.WithSelector("#"+errorID), datastar.WithModeOuter())
	_ = sse.MarshalAndPatchSignals(map[string]bool{"ticketSaving": false, "ticketsLoading": false})
}

func (c *Controller) ticketOperationSuccess(w http.ResponseWriter, r *http.Request, closeDialog bool) {
	data := c.ticketsTableData(r)
	var body strings.Builder
	_ = TicketsTableRegion(data).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#tickets-table-region"), datastar.WithModeOuter())
	_ = sse.MarshalAndPatchSignals(map[string]bool{"ticketSaving": false, "ticketsLoading": false})
	if closeDialog {
		_ = sse.ExecuteScript("window.templ?.dialog?.close('ticket-dialog');")
	}
}
