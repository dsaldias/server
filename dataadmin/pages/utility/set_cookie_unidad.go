package utility

import "net/http"

func SetUnidadOnCookie(unidadid string, w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "galletita_traviesa_unidad_default",
		Value:    unidadid,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,                  // true en producción con HTTPS
		SameSite: http.SameSiteNoneMode, // front y back en dominios diferentes
	})
}
