package xnotificaciones

/* import (
	"fmt"
	"github.com/dsaldias/server/graph_auth/model"
	"sync"
)

type Chan struct {
	mu            sync.Mutex
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
	channels := c.subscriptores[userID]
	for i, subscriber := range channels {
		if subscriber == ch {
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
	c.mu.Lock()
	defer c.mu.Unlock()

	chans := c.subscriptores[userid]
	out := make([]chan *model.XNotificacion, len(chans))
	copy(out, chans)
	return out
}

func (c *Chan) TotalConectados() (int, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.totalSubs, len(c.subscriptores)
}

func (c *Chan) IdsConectados() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	conects := []string{}
	for id := range c.subscriptores {
		conects = append(conects, id)
	}
	return conects
}

func (c *Chan) Broadcast(xn *model.XNotificacion) {
	c.mu.Lock()
	subs := make(map[string][]chan *model.XNotificacion)
	for k, v := range c.subscriptores {
		subs[k] = append([]chan *model.XNotificacion{}, v...)
	}
	c.mu.Unlock()

	for userID, channels := range subs {
		for _, ch := range channels {
			select {
			case ch <- xn:
			default:
				fmt.Printf("Notification channel for user %s is full.\n", userID)
			}
		}
	}
}

func (c *Chan) SSEBroadcast(xn *model.XNotificacion) {
	sseManager.Broadcast(xn.DataJSON)
}
*/

// ORIGINAL consumia arto ram
/* package xnotificaciones

import (
	"fmt"
	"github.com/dsaldias/server/graph_auth/model"
	"sync"
)

type Chan struct {
	mu            sync.Mutex
	subscriptores map[string][]chan *model.XNotificacion
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
}

func (c *Chan) RemoveSubscriber(userID string, ch chan *model.XNotificacion) {
	c.mu.Lock()
	defer c.mu.Unlock()
	channels := c.subscriptores[userID]
	for i, subscriber := range channels {
		if subscriber == ch {
			c.subscriptores[userID] = append(channels[:i], channels[i+1:]...)
			break
		}
	}
	if len(c.subscriptores[userID]) == 0 {
		delete(c.subscriptores, userID)
	}
}

func (c *Chan) GetSubsByUser(userid string) []chan *model.XNotificacion {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.subscriptores[userid]
}

func (c *Chan) TotalConectados() (int, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	total := 0
	sub := len(c.subscriptores)
	for _, lista := range c.subscriptores {
		total += len(lista)
	}
	return total, sub
}

func (c *Chan) IdsConectados() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	conects := []string{}
	for id := range c.subscriptores {
		conects = append(conects, id)
	}
	return conects
}

func (c *Chan) Broadcast(xn *model.XNotificacion) {
	c.mu.Lock()
	defer c.mu.Unlock()
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
	sseManager.Broadcast(xn.DataJSON)
}

func (c *Chan) SSEBroadcast(xn *model.XNotificacion) {
	sseManager.Broadcast(xn.DataJSON)
}
*/
