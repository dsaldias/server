package xnotificaciones

import (
	"fmt"
	"sync"

	"github.com/dsaldias/server/graph_auth/model"
)

type Chan struct {
	mu            sync.RWMutex
	subscriptores map[string][]chan *model.XNotificacion
	totalSubs     int
}

var global *Chan
var once sync.Once
var sseManager = NewSSEManager(10)

func InitializeGlobal() {
	once.Do(func() {
		global = &Chan{
			subscriptores: make(map[string][]chan *model.XNotificacion),
		}
	})
}

func GetGlobal() *Chan {
	if global == nil {
		InitializeGlobal()
	}
	return global
}

func (c *Chan) AddSubscriber(userID string, ch chan *model.XNotificacion) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscriptores[userID] = append(c.subscriptores[userID], ch)
	c.totalSubs++
}

func (c *Chan) RemoveSubscriber(userID string, ch chan *model.XNotificacion) {
	c.mu.Lock()
	defer c.mu.Unlock()
	channels, ok := c.subscriptores[userID]
	if !ok {
		return
	}
	for i, subscriber := range channels {
		if subscriber == ch {
			channels[i] = nil
			c.subscriptores[userID] = append(channels[:i], channels[i+1:]...)
			c.totalSubs--
			break
		}
	}
	if len(c.subscriptores[userID]) == 0 {
		delete(c.subscriptores, userID)
	}
}

func (c *Chan) GetSubsByUser(userid string) []chan *model.XNotificacion {
	c.mu.RLock() // Usa RLock, es solo lectura
	defer c.mu.RUnlock()

	original, ok := c.subscriptores[userid]
	if !ok {
		return nil
	}

	// Crear una copia segura para el que llama
	out := make([]chan *model.XNotificacion, len(original))
	copy(out, original)
	return out
}

func (c *Chan) TotalConectados() (int, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.totalSubs, len(c.subscriptores)
}

func (c *Chan) IdsConectados() []string {
	c.mu.RLock() // Cambiar Lock por RLock
	defer c.mu.RUnlock()

	conects := make([]string, 0, len(c.subscriptores)) // Pre-asignar capacidad es más rápido
	for id := range c.subscriptores {
		conects = append(conects, id)
	}
	return conects
}

func (c *Chan) Broadcast(xn *model.XNotificacion) {
	c.mu.RLock()
	// defer c.mu.RUnlock()
	for userID, channels := range c.subscriptores {

		for _, ch := range channels {
			select {
			case ch <- xn:
				// Successfully sent the notification
			default:
				// Channel is full, consider logging this event
				fmt.Printf("Notification channel for user %s is full.\n", userID)
			}
		}
	}
	c.mu.RUnlock()

	sseManager.Broadcast(xn.DataJSON)
}

func (c *Chan) SSEBroadcast(xn *model.XNotificacion) {
	sseManager.Broadcast(xn.DataJSON)
}
