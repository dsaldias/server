package utility

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	PAGDefault = 10
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

func GetPaginacionParams(r *http.Request) (int32, int32) {
	xpagina := r.URL.Query().Get("page")
	xtam := r.URL.Query().Get("size")

	pagina := int32(1)
	tam := int32(PAGDefault)
	pag, err := strconv.ParseInt(xpagina, 10, 32)
	if err == nil {
		pagina = int32(pag)
	}
	siz, err := strconv.ParseInt(xtam, 10, 32)
	if err == nil {
		tam = int32(siz)
	}

	return pagina, tam
}
func GetPaginacion(r *http.Request, paginas int32, targetHtmx string) MiPaginacion {
	xpagina := r.URL.Query().Get("page")
	xtam := r.URL.Query().Get("size")

	pagina := int32(1)
	tam := int32(PAGDefault)
	pag, err := strconv.ParseInt(xpagina, 10, 32)
	if err == nil {
		pagina = int32(pag)
	}
	siz, err := strconv.ParseInt(xtam, 10, 32)
	if err == nil {
		tam = int32(siz)
	}

	pagi := MiPaginacion{
		Paginas:    int(paginas),
		Pagina:     int(pagina),
		Size:       int(tam),
		TargetHtmx: targetHtmx,
	}

	return pagi
}

func generarPaginas(p MiPaginacion) []ItemPaginacion {
	if p.Paginas <= 0 {
		return nil
	}

	resultado := []ItemPaginacion{}

	incluir := func(n int) {
		resultado = append(resultado, ItemPaginacion{
			Pagina:   n,
			EsActual: n == p.Pagina,
		})
	}

	incluir(1)

	inicio := p.Pagina - 2
	fin := p.Pagina + 2

	if inicio < 2 {
		inicio = 2
	}

	if fin > p.Paginas-1 {
		fin = p.Paginas - 1
	}

	if inicio > 2 {
		resultado = append(resultado, ItemPaginacion{
			EsSalto: true,
		})
	}

	for i := inicio; i <= fin; i++ {
		incluir(i)
	}

	if fin < p.Paginas-1 {
		resultado = append(resultado, ItemPaginacion{
			EsSalto: true,
		})
	}

	if p.Paginas > 1 {
		incluir(p.Paginas)
	}

	return resultado
}

func paginaURL(pagina, size int, tab string) string {
	return fmt.Sprintf(
		"?page=%d&size=%d&tab=%s",
		pagina,
		size,
		tab,
	)
}

func isSorted(props MiTablaProps) string {
	if props.Ordenable {
		return "true"
	}
	return "false"
}

func IsOnlyHtmx(r *http.Request) bool {
	xpagina := r.URL.Query().Get("page")
	xref := r.URL.Query().Get("xrefresh")

	if xref != "" || xpagina != "" {
		if r.Header.Get("HX-Request") == "true" {
			return true
		}
	}
	return false
}
