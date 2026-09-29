package utility

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utils"
)

func Is_Auth(
	db *sql.DB,
	w http.ResponseWriter,
	r *http.Request,
	metodo_name string,
) (*utils.AuthData, error) {

	xauth, err := utils.CtxValue(r.Context(), db, metodo_name)
	if err != nil {
		if r.Header.Get("HX-Request") == "true" {
			ErrorResponse(w, r, err, nil)
			return nil, err
		}

		isIframe := r.URL.Query().Get("iframe") == "true"
		enlace_principal := "/adminx/"
		if isIframe {
			enlace_principal += "?iframe=true"
		}

		http.Redirect(w, r, enlace_principal, http.StatusSeeOther)
		return nil, err
	}
	return xauth, nil
}
