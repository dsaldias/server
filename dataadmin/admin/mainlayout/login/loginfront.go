package xlogin

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataadmin/admin/utility"
	"github.com/dsaldias/server/dataadmin/admin/utils"
	"github.com/dsaldias/server/dataauth/login"
	"github.com/dsaldias/server/graph_auth/model"
)

type Logincontroller struct {
	DB *sql.DB
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
			token_portal, err2 := LoginPortal(usuario, clave)
			if err2 != nil {
				if xis_relogin {
					http.Error(w, err2.Error(), http.StatusForbidden)
				} else {
					utils.ErrorTpl(err2.Error()).Render(r.Context(), w)
				}
				return
			}

			if len(token_portal) == 0 {
				t := "datos del portal incorrectos"
				if xis_relogin {
					http.Error(w, t, http.StatusForbidden)
				} else {
					utils.ErrorTpl(t).Render(r.Context(), w)
				}
				return
			}

			res, err3 := GetMe(token_portal)
			if err3 != nil {
				if xis_relogin {
					http.Error(w, err3.Error(), http.StatusForbidden)
				} else {
					utils.ErrorTpl(err3.Error()).Render(r.Context(), w)
				}
				return
			}

			e := GuardarDatosPortal(c.DB, usuario, clave, res)
			if e != nil {
				if xis_relogin {
					http.Error(w, e.Error(), http.StatusForbidden)
				} else {
					utils.ErrorTpl(e.Error()).Render(r.Context(), w)
				}
				return
			}

			newlogin, err4 := login.Login2(r.Context(), c.DB, data)
			if err4 != nil {
				if xis_relogin {
					http.Error(w, err4.Error(), http.StatusForbidden)
				} else {
					utils.ErrorTpl(err4.Error()).Render(r.Context(), w)
				}
				return
			}
			logindata = newlogin
		}

		if logindata == nil {
			if xis_relogin {
				http.Error(w, "datos sistema y portal incorrectos", http.StatusForbidden)
			} else {
				utils.ErrorTpl("datos sistema y portal incorrectos").Render(r.Context(), w)
			}
			return
		}

		if len(logindata.Me.Roles) == 0 {
			if xis_relogin {
				http.Error(w, "datos correctos, pero no tienes ningun rol asignado.", http.StatusForbidden)
			} else {
				utils.ErrorTpl("datos correctos, pero no tienes ningun rol asignado.").Render(r.Context(), w)
			}
			return
		}

		unidad := logindata.Me.Roles[0].Unidad
		utility.SetUnidadOnCookie(unidad.ID, w, r)

		/* http.SetCookie(w, &http.Cookie{
			Name:     "galletita_traviesa_unidad_default",
			Value:    unidad.ID,
			Path:     "/",
			HttpOnly: true,
			Secure:   true, // true en producción con HTTPS
			// SameSite: http.SameSiteLaxMode,
			SameSite: http.SameSiteNoneMode, // front y back en dominios diferentes
		}) */

		if !xis_relogin {
			w.Header().Set("HX-Redirect", "/webx/main")
		}

		w.WriteHeader(http.StatusNoContent)

	})
}
