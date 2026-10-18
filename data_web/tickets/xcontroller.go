package tickets

import (
	"database/sql"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type Controller struct{ DB *sql.DB }

type TicketsPageData struct {
	Listado []*model.RespTickets
	Estado  string
	Error   string
}

type ticketSignals struct {
	Estado string `json:"ticketEstado"`
}

func (c *Controller) TicketsRender(w http.ResponseWriter, r *http.Request) {
	data := c.ticketsTableData(r)
	_ = RenderTickets(data).Render(r.Context(), w)
}

func (c *Controller) TicketsListaHTML(w http.ResponseWriter, r *http.Request) {
	data := c.ticketsTableData(r)
	if r.Header.Get("Datastar-Request") != "true" {
		_ = TicketsTableRegion(data).Render(r.Context(), w)
		return
	}

	var signals ticketSignals
	if err := datastar.ReadSignals(r, &signals); err == nil {
		data.Estado = strings.TrimSpace(signals.Estado)
		data.Listado, _ = c.listTickets(r, data.Estado)
		if data.Listado == nil {
			data.Error = "No fue posible cargar los tickets."
		}
	}
	c.ticketsPatch(w, r, data)
}

func (c *Controller) ticketsTableData(r *http.Request) TicketsPageData {
	estado := strings.TrimSpace(r.URL.Query().Get("estado"))
	data := TicketsPageData{Estado: estado}
	data.Listado, _ = c.listTickets(r, estado)
	if data.Listado == nil {
		data.Error = "No fue posible cargar los tickets."
	}
	return data
}

func (c *Controller) listTickets(r *http.Request, estado string) ([]*model.RespTickets, error) {
	listado, err := xrepo.AllTickets(r.Context(), c.DB, model.QueryTickets{Estado: optionalString(estado)})
	if err != nil {
		return nil, err
	}
	if estado == "" {
		return listado, nil
	}
	filtered := make([]*model.RespTickets, 0, len(listado))
	for _, ticket := range listado {
		if ticket != nil && strings.EqualFold(ticket.Estado, estado) {
			filtered = append(filtered, ticket)
		}
	}
	return filtered, nil
}

func (c *Controller) ticketsPatch(w http.ResponseWriter, r *http.Request, data TicketsPageData) {
	var body strings.Builder
	_ = TicketsTableRegion(data).Render(r.Context(), &body)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(body.String(), datastar.WithSelector("#tickets-table-region"), datastar.WithModeOuter())
	_ = sse.ExecuteScript("history.replaceState({},''," + strconv.Quote(ticketQueryURL(data)) + ")")
	_ = sse.MarshalAndPatchSignals(map[string]any{"ticketsLoading": false})
}

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

func ticketsRefreshScript() string {
	return "$ticketsLoading=true; @get('" + ADMIN_TICKETS_LISTAHTML + "')"
}

func ticketsStatusScript() string {
	return "$ticketsLoading=true; @get('" + ADMIN_TICKETS_LISTAHTML + "')"
}

func ticketsLoadingEffectScript() string {
	return "const button=document.querySelector('[data-tickets-refresh]');if(button){button.disabled=$ticketsLoading;if($ticketsLoading){button.setAttribute('aria-busy','true');button.querySelector('[data-tickets-refresh-icon]')?.classList.add('hidden');const label=button.querySelector('[data-tickets-refresh-label]');if(label)label.textContent='Cargando...';}else{button.removeAttribute('aria-busy');button.querySelector('[data-tickets-refresh-icon]')?.classList.remove('hidden');const label=button.querySelector('[data-tickets-refresh-label]');if(label)label.textContent='Refrescar';}}"
}
