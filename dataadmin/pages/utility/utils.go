package utility

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func TemplUIJS(templuiPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		file := filepath.Base(r.URL.Path)

		var found string

		err := filepath.Walk(
			templuiPath,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !info.IsDir() && info.Name() == file {
					found = path
					return filepath.SkipDir
				}

				return nil
			},
		)

		if err != nil || found == "" {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, found)
	})
}

func ParseBodyToJSON(r *http.Request) (map[string]any, error) {
	var data map[string]any

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return nil, err
	}

	if value, ok := data["fecha_solicitud"].(string); ok {
		fechaSolicitud, err := time.Parse(
			"2006-01-02T15:04",
			value,
		)
		if err != nil {
			return nil, err
		}

		data["fecha_solicitud"] = fechaSolicitud.Format(time.RFC3339)
	}

	if value, ok := data["sede_id"].(string); ok {
		sedeID, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, err
		}

		data["sede_id"] = sedeID
	}

	if value, ok := data["solicitante_id"].(string); ok {
		solicitanteID, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, err
		}

		data["solicitante_id"] = solicitanteID
	}

	return data, nil
}

func ResponseOkJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ErrorResponse(w http.ResponseWriter, r *http.Request, err error, status *int) {
	if status != nil {
		http.Error(w, err.Error(), *status)
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

/* func RenderPage(
	w http.ResponseWriter,
	r *http.Request,
	layout *mainlayout.MainController,
	contenido templ.Component,
) {
	if r.Header.Get("HX-Request") == "true" {
		contenido.Render(r.Context(), w)
		return
	}

	layout.RenderLayout(w, r, contenido)
} */
