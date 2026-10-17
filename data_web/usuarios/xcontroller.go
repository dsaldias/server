package usuarios

import (
	"database/sql"
	"net/http"
)

type Controller struct {
	DB *sql.DB
}

func (c *Controller) UsuariosRender(w http.ResponseWriter, r *http.Request) {
	/* if c.authenticated(r) {
		http.Redirect(w, r, ADMIN_USUARIOS, http.StatusSeeOther)
		return
	} */

	RenderUsuarios().Render(r.Context(), w)
}
