package avisos

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dsaldias/server/dataauth/xrepo"
	"github.com/dsaldias/server/graph_auth/model"
	"github.com/starfederation/datastar-go/datastar"
)

type Controller struct{ DB *sql.DB }

type AvisosPageData struct {
	Listado []*model.Notificacion
	Error   string
}

func (c *Controller) AvisosRender(w http.ResponseWriter, r *http.Request) {
	data := c.avisosTableData(r)
	_ = RenderAvisos(data).Render(r.Context(), w)
}

func (c *Controller) AvisosListaHTML(w http.ResponseWriter, r *http.Request) {
	data := c.avisosTableData(r)
	if r.Header.Get("Datastar-Request") == "true" {
		var body strings.Builder
		_ = AvisosTableRegion(data).Render(r.Context(), &body)
		sse := datastar.NewSSE(w, r)
		_ = sse.PatchElements(body.String(), datastar.WithSelector("#avisos-table-region"), datastar.WithModeOuter())
		_ = sse.MarshalAndPatchSignals(map[string]any{"avisosLoading": false})
		return
	}
	_ = AvisosTableRegion(data).Render(r.Context(), w)
}

func (c *Controller) avisosTableData(r *http.Request) AvisosPageData {
	data := AvisosPageData{}
	data.Listado, _ = xrepo.Notificaciones(r.Context(), c.DB)
	if data.Listado == nil {
		data.Error = "No fue posible cargar los avisos activos."
	}
	return data
}

func avisosRefreshScript() string {
	return "$avisosLoading=true; @get('" + ADMIN_AVISOS_LISTAHTML + "')"
}

func avisosLoadingEffectScript() string {
	return "const button=document.querySelector('[data-avisos-refresh]');if(button){button.disabled=$avisosLoading;if($avisosLoading){button.setAttribute('aria-busy','true');button.querySelector('[data-avisos-refresh-icon]')?.classList.add('hidden');const label=button.querySelector('[data-avisos-refresh-label]');if(label)label.textContent='Cargando...';}else{button.removeAttribute('aria-busy');button.querySelector('[data-avisos-refresh-icon]')?.classList.remove('hidden');const label=button.querySelector('[data-avisos-refresh-label]');if(label)label.textContent='Refrescar';}}"
}

func avisoDate(value time.Time) string {
	if value.IsZero() {
		return "—"
	}
	return value.UTC().Format("02/01/2006 15:04")
}

func avisoDateISO(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("2006-01-02T15:04")
}

func avisoSignalString(value string) string { return strconv.Quote(value) }
