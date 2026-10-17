package tickets

import "net/http"

type Controller struct{}

func (c *Controller) TicketsRender(w http.ResponseWriter, r *http.Request) {
	_ = RenderTickets().Render(r.Context(), w)
}
