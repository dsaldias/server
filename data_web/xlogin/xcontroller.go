package xlogin

import (
	"net/http"
)

type Controller struct {
}

func (c *Controller) LoginRender(w http.ResponseWriter, r *http.Request) {
	MainLayout("home", Default()).Render(r.Context(), w)
}
