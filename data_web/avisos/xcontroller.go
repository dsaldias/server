package avisos

import "net/http"

type Controller struct{}

func (c *Controller) NotificacionesRender(w http.ResponseWriter, r *http.Request) {
	_ = RenderNotificaciones().Render(r.Context(), w)
}
