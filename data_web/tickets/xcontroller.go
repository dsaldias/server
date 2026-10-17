package tickets

import (
	"database/sql"
	"html"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type Controller struct{ DB *sql.DB }

type TicketsPageData struct {
	Listado  []*model.RespTickets
	Detalles map[string]*model.Ticket
	Estado   string
	Error    string
}

type ticketSignals struct {
	TicketEstado string
}

type ticketFormSignals struct {
	Problema     string
	Respuesta    string
	TicketID     string
	TicketAction string
}

func (c *Controller) TicketsRender(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		c.ticketSave(w, r)
		return
	}
	data := c.ticketsPageData(r)
	if r.Header.Get("Datastar-Request") == "true" {
		c.ticketsPatch(w, r, data)
		return
	}
	_ = RenderTickets(data).Render(r.Context(), w)
}

func (c *Controller) ticketsPageData(r *http.Request) TicketsPageData {
	return c.ticketsPageDataWithState(r, strings.TrimSpace(r.URL.Query().Get("estado")))
}

func (c *Controller) ticketsPageDataWithState(r *http.Request, estado string) TicketsPageData {
	data := TicketsPageData{Estado: estado, Detalles: make(map[string]*model.Ticket)}
	listado, err := xrepo.AllTickets(r.Context(), c.DB, model.QueryTickets{Estado: optionalString(estado)})
	if err != nil {
		data.Error = "No fue posible cargar los tickets."
		return data
	}
	for _, ticket := range listado {
		if ticket == nil || (estado != "" && !strings.EqualFold(ticket.Estado, estado)) {
			continue
		}
		data.Listado = append(data.Listado, ticket)
		if detail, detailErr := xrepo.VerTicketConRespuestas(r.Context(), c.DB, ticket.ID); detailErr == nil && detail != nil {
			data.Detalles[ticket.ID] = detail
		}
	}
	return data
}

func (c *Controller) ticketsPatch(w http.ResponseWriter, r *http.Request, data TicketsPageData) {
	var signals ticketSignals
	if err := datastar.ReadSignals(r, &signals); err == nil {
		data = c.ticketsPageDataWithState(r, strings.TrimSpace(signals.TicketEstado))
	}
	var body strings.Builder
	_ = TicketsTableRegion(data).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#tickets-table-region"), datastar.WithModeOuter())
	_ = sse.ExecuteScript("history.replaceState({},''," + strconv.Quote(ticketQueryURL(data)) + ")")
}

func (c *Controller) ticketSave(w http.ResponseWriter, r *http.Request) {
	var signals ticketFormSignals
	if err := datastar.ReadSignals(r, &signals); err != nil {
		_ = r.ParseForm()
		signals.Problema, signals.Respuesta = r.FormValue("problema"), r.FormValue("respuesta")
		signals.TicketID, signals.TicketAction = r.FormValue("ticketID"), r.FormValue("ticketAction")
	}
	if signals.TicketAction == "" {
		signals.TicketAction = r.URL.Query().Get("action")
	}
	var err error
	switch signals.TicketAction {
	case "reply":
		if strings.TrimSpace(signals.TicketID) == "" || strings.TrimSpace(signals.Respuesta) == "" {
			err = ticketValidationError("La respuesta es obligatoria.")
		} else {
			_, err = xrepo.UpdateTicket(r.Context(), c.DB, model.NewTicketRespuesta{TicketsID: signals.TicketID, Respuesta: strings.TrimSpace(signals.Respuesta)})
		}
	case "close":
		_, err = xrepo.CerrarTicket(r.Context(), c.DB, strings.TrimSpace(signals.TicketID))
	default:
		if strings.TrimSpace(signals.Problema) == "" {
			err = ticketValidationError("El problema es obligatorio.")
		} else {
			_, err = xrepo.CreateTicket(r.Context(), c.DB, model.NewTicket{Problema: strings.TrimSpace(signals.Problema)})
		}
	}
	sse := datastar.NewSSE(w, r)
	if err != nil {
		selector := "#ticket-form-error"
		if signals.TicketAction == "reply" {
			selector = "#ticket-response-form-error"
		}
		errorID := "ticket-form-error"
		if signals.TicketAction == "reply" {
			errorID = "ticket-response-form-error"
		}
		_ = sse.PatchElements(ticketFormErrorWithID(errorID, err.Error()), datastar.WithSelector(selector), datastar.WithModeOuter())
		_ = sse.MarshalAndPatchSignals(map[string]bool{"ticketSaving": false, "ticketSaved": false})
		return
	}
	data := c.ticketsPageData(r)
	var body strings.Builder
	_ = TicketsTableRegion(data).Render(r.Context(), &body)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#tickets-table-region"), datastar.WithModeOuter())
	_ = sse.MarshalAndPatchSignals(map[string]bool{"ticketSaving": false, "ticketSaved": true})
	if (signals.TicketAction == "reply" || signals.TicketAction == "close") && strings.TrimSpace(signals.TicketID) != "" {
		_ = sse.ExecuteScript("window.templ?.dialog?.close(" + strconv.Quote("ticket-view-"+strings.TrimSpace(signals.TicketID)) + ")")
	}
}

type ticketValidationError string

func (e ticketValidationError) Error() string { return string(e) }

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func ticketSignalString(value string) string { return strconv.Quote(value) }

func ticketQueryURL(data TicketsPageData) string {
	if data.Estado == "" {
		return ADMIN_TICKETS
	}
	return ADMIN_TICKETS + "?estado=" + url.QueryEscape(data.Estado)
}

func ticketDate(value time.Time) string {
	if value.IsZero() {
		return "—"
	}
	return value.Local().Format("02/01/2006 15:04")
}

func ticketText(value *string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return "Sin respuesta"
	}
	return *value
}

func ticketStatusClass(value string) string {
	switch strings.ToLower(value) {
	case "cerrado":
		return "bg-muted text-muted-foreground"
	case "soporte":
		return "bg-blue-500/10 text-blue-700 dark:text-blue-300"
	case "cliente":
		return "bg-amber-500/10 text-amber-700 dark:text-amber-300"
	default:
		return "bg-emerald-500/10 text-emerald-700 dark:text-emerald-300"
	}
}

func ticketViewAttributes(id string) templ.Attributes {
	return templ.Attributes{"type": "button", "data-slot": "dialog-trigger", "data-base-ui-click-trigger": true, "data-templ-controls": "ticket-view-" + id, "aria-haspopup": "dialog", "aria-expanded": "false"}
}

func ticketCloseAttributes(id string) templ.Attributes {
	script := "if(window.confirm(" + strconv.Quote("¿Deseas cerrar este ticket? Esta acción no se puede deshacer.") + ")){ $ticketID=" + strconv.Quote(id) + "; $ticketSaving=true; $ticketAction='close'; @post('/adminx/tickets?action=close') }"
	return templ.Attributes{"type": "button", "data-on:click": script}
}

func ticketFormError(message string) string {
	return ticketFormErrorWithID("ticket-form-error", message)
}

func ticketFormErrorWithID(id, message string) string {
	return "<div id=\"" + html.EscapeString(id) + "\" role=\"alert\" class=\"rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive\">" + html.EscapeString(message) + "</div>"
}
