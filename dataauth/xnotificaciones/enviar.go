package xnotificaciones

import (
	"context"
	"fmt"

	"github.com/dsaldias/server/graph_auth/model"
)

func EnviarNotificacion(ctx context.Context, titulo string, datos *DataNotify) (bool, error) {
	cha := GetGlobal()

	m := ""
	if datos != nil {
		if s, err := formatToJson(*datos); err == nil {
			m = s
		} else {
			fmt.Println("ERROR NOTIFY: ", err.Error())
		}
	}

	xn := &model.XNotificacion{
		Title:    titulo,
		DataJSON: m,
	}

	cha.Broadcast(xn)

	return true, nil
}

func EnviarSSENotificacion(ctx context.Context, titulo string, datos *DataNotify) (bool, error) {
	cha := GetGlobal()

	m := ""
	if datos != nil {
		if s, err := formatToJson(*datos); err == nil {
			m = s
		} else {
			fmt.Println("ERROR NOTIFY SSE: ", err.Error())
		}
	}

	xn := &model.XNotificacion{
		Title:    titulo,
		DataJSON: m,
	}

	cha.SSEBroadcast(xn)

	return true, nil
}
