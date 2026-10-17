package roles

import "net/http"

type Controller struct{}

func (c *Controller) RolesRender(w http.ResponseWriter, r *http.Request) {
	_ = RenderRoles().Render(r.Context(), w)
}
