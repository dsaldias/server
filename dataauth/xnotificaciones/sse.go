package xnotificaciones

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
)

type SSEManager struct {
	clients    map[string]map[chan string]struct{}
	mutex      sync.RWMutex
	bufferSize int
	metrics    Metrics
}

type Metrics struct {
	Sent    int64
	Skipped int64
}

func NewSSEManager(bufferSize int) *SSEManager {
	return &SSEManager{
		clients:    make(map[string]map[chan string]struct{}),
		bufferSize: bufferSize,
	}
}

func (m *SSEManager) AddClient(userID string) chan string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	messageChan := make(chan string, m.bufferSize)

	if _, exists := m.clients[userID]; !exists {
		m.clients[userID] = make(map[chan string]struct{})
	}

	m.clients[userID][messageChan] = struct{}{}
	return messageChan
}

func (m *SSEManager) RemoveClient(userID string, messageChan chan string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if userClients, exists := m.clients[userID]; exists {
		if _, ok := userClients[messageChan]; ok {
			delete(userClients, messageChan)
			close(messageChan)

			if len(userClients) == 0 {
				delete(m.clients, userID)
			}
			log.Printf("Cliente removido: %s", userID)
		}
	}
}

func (m *SSEManager) Broadcast(message string) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, userClients := range m.clients {
		for client := range userClients {
			if safeSend(client, message) {
				atomic.AddInt64(&m.metrics.Sent, 1)
			} else {
				atomic.AddInt64(&m.metrics.Skipped, 1)
			}
			// select {
			// case client <- message:
			// 	totalSent++
			// default:
			// 	log.Printf("Canal lleno para el usuario %s, omitiendo mensaje", userID)
			// 	totalSkipped++
			// }
		}
	}

}

func safeSend(ch chan<- string, msg string) bool {
	select {
	case ch <- msg:
		return true
	default:
		return false
	}
}

func SSEHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userid")
	if userID == "" {
		http.Error(w, "No autorizado, falta el userid", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE no soportado", http.StatusInternalServerError)
		return
	}

	messageChan := sseManager.AddClient(userID)
	defer sseManager.RemoveClient(userID, messageChan)

	fmt.Fprintf(w, "event: connected\ndata: %s\n\n", "Conexión SSE establecida")
	flusher.Flush()

	// Mantener conexión abierta
	for {
		select {
		case msg := <-messageChan:
			// Enviar mensaje al cliente
			fmt.Fprintf(w, "event: notification\ndata: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return // Cerrar cuando el cliente se desconecte
		}
	}
}
