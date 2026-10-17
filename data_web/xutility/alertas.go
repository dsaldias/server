package xutility

import (
	"strconv"
	"time"

	"github.com/dsaldias/server/graph_auth/model"
)

func activeAlertCount(alerts []*model.Notificacion) string {
	return strconv.Itoa(len(alerts))
}

func activeAlertDate(value time.Time) string {
	if value.IsZero() {
		return "—"
	}
	return value.UTC().Format("02/01/2006 15:04")
}
