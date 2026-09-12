package usuarios

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataadmin/admin/mainlayout"
	"github.com/dsaldias/server/dataadmin/admin/mainlayout/principal"
)

type TestController struct {
	DB *sql.DB
	C  *mainlayout.MainController
}

func (c *TestController) Listar(w http.ResponseWriter, r *http.Request) {
	ini := principal.Inicio()
	c.C.RenderPage(w, r, ini)
}
