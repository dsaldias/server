package utils

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type clientInfo struct {
	count    int
	lastSeen time.Time
}

type shard struct {
	clients map[string]*clientInfo
	mu      sync.Mutex
}

type RateLimiter struct {
	shards []*shard      // Shards para reducir contención
	limit  int           // Límite de solicitudes
	window time.Duration // Ventana de tiempo
	stop   chan struct{} // Canal para detener limpieza
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	shards := 256
	rl := &RateLimiter{
		shards: make([]*shard, shards),
		limit:  limit,
		window: window,
		stop:   make(chan struct{}),
	}
	for i := range rl.shards {
		rl.shards[i] = &shard{
			clients: make(map[string]*clientInfo),
		}
	}
	rl.startCleanup(time.Minute) // Ajusta según necesidad
	return rl
}

func (rl *RateLimiter) RateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := getClientKey(r) // Implementa esta función de forma segura

		// Selecciona un shard usando hash del clientIP
		shardIndex := int(hash(clientIP) % uint32(len(rl.shards))) // Conversión segura
		shard := rl.shards[shardIndex]

		shard.mu.Lock()
		defer shard.mu.Unlock()

		info, exists := shard.clients[clientIP]
		if !exists {
			info = &clientInfo{
				count:    0,
				lastSeen: time.Now(),
			}
			shard.clients[clientIP] = info
		}

		now := time.Now()
		if now.Sub(info.lastSeen) > rl.window {
			info.count = 0
			info.lastSeen = now
		}

		if info.count >= rl.limit {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		info.count++
		info.lastSeen = now
		next.ServeHTTP(w, r)
	})
}

// Limpieza periódica en cada shard
func (rl *RateLimiter) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, shard := range rl.shards {
					shard.mu.Lock()
					now := time.Now()
					for ip, info := range shard.clients {
						if now.Sub(info.lastSeen) > rl.window*2 {
							delete(shard.clients, ip)
						}
					}
					shard.mu.Unlock()
				}
			case <-rl.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

// Funciones auxiliares (implementa según tu lógica)
func getClientKey(r *http.Request) string {
	// Ejemplo: Combinar IP y SESSIONKEY para mayor seguridad
	ip := clientIP(r)
	sessionKey := r.Header.Get("SESSIONKEY")
	return ip + "|" + sessionKey
}

func hash(s string) uint32 {
	h := uint32(2166136261) // Usar uint32 en lugar de int
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

func clientIP(r *http.Request) string {
	// Prioriza X-Forwarded-For si existe (primer IP)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
