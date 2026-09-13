package xlogin

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataadmin/pages/utility"
	"github.com/dsaldias/server/dataauth/login"
	"github.com/dsaldias/server/graph_auth/model"
)

type Logincontroller struct {
	DB *sql.DB
}

func (c *Logincontroller) Logout(w http.ResponseWriter, r *http.Request) {
	cookies := []string{
		"galletita_traviesa",
	}

	for _, name := range cookies {
		http.SetCookie(w, &http.Cookie{
			Name:   name,
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
	}

	http.Redirect(w, r, "/adminx/", http.StatusSeeOther)
}

func (c *Logincontroller) Login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		usuario := r.FormValue("usuario")
		clave := r.FormValue("clave")
		xis_relogin := r.FormValue("xis_relogin") == "true"

		data := model.NewLogin2{
			Username: usuario,
			Password: clave,
		}
		logindata, err := login.Login2(r.Context(), c.DB, data)
		if err != nil {
			utility.ErrorTpl(err.Error()).Render(r.Context(), w)
			return
		}

		if logindata == nil {
			if xis_relogin {
				http.Error(w, "datos sistema y portal incorrectos", http.StatusForbidden)
			} else {
				utility.ErrorTpl("datos sistema y portal incorrectos").Render(r.Context(), w)
			}
			return
		}

		if len(logindata.Me.Roles) == 0 {
			t := "datos correctos, pero no tienes ningun rol asignado."
			if xis_relogin {
				http.Error(w, t, http.StatusForbidden)
			} else {
				utility.ErrorTpl(t).Render(r.Context(), w)
			}
			return
		}

		if !xis_relogin {
			w.Header().Set("HX-Redirect", "/adminx/main")
		}

		w.WriteHeader(http.StatusNoContent)

	})
}
