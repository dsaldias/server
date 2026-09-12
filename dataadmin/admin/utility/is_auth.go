package utility

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utils"
)

func Is_Auth(db *sql.DB, w http.ResponseWriter, r *http.Request, metodo_name string) (*utils.AuthData, error) {
	xauth, err := utils.CtxValue(r.Context(), db, metodo_name)
	if err != nil {
		// fmt.Println("errrrrolesss", err)

		if r.Header.Get("HX-Request") == "true" {
			ErrorResponse(w, r, err, nil)
			return nil, err
		}

		http.Redirect(w, r, "/webx/", http.StatusSeeOther)
		return nil, err
	}
	return xauth, nil
}
