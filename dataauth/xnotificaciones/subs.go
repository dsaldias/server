package xnotificaciones

import (
	"context"
	"fmt"

	"github.com/dsaldias/server/graph_auth/model"
)

func NotificacionesSubs(ctx context.Context, userid string) (<-chan *model.XNotificacion, error) {

	cha := GetGlobal()

	ch := make(chan *model.XNotificacion, 10)
	cha.AddSubscriber(userid, ch)

	go func() {
		<-ctx.Done()
		cha.RemoveSubscriber(userid, ch)
		close(ch)

		notificarConectados(ctx, cha)
	}()

	notificarConectados(ctx, cha)

	return ch, nil
}

func notificarConectados(ctx context.Context, cha *Chan) {

	total_conectados, conectados := cha.TotalConectados()
	tipo := "conectados"
	mapa := map[string]int{}
	mapa["total_conectados"] = total_conectados
	mapa["conectados"] = conectados

	d := DataNotify{Tipo: &tipo, Datos: mapa}
	EnviarNotificacion(ctx, fmt.Sprintf("%d", total_conectados), &d)

}
