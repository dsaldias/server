package unidades

import "net/http"

type Controller struct{}

func (c *Controller) UnidadesRender(w http.ResponseWriter, r *http.Request) {
	_ = RenderUnidades().Render(r.Context(), w)
}
