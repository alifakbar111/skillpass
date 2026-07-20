package attendance

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// upgrader is configured at startup via SetAllowedOrigins. The default
// allows only localhost dev origins; production deployments MUST set
// this via cfg.CORSOrigin (see cmd/server/main.go).
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // non-browser clients (no Origin header) are allowed
		}
		for _, o := range allowedOrigins {
			if strings.EqualFold(o, origin) {
				return true
			}
		}
		return false
	},
}

var allowedOrigins = []string{
	"http://localhost:4200",
	"https://localhost:4200",
	"http://127.0.0.1:4200",
}

// SetAllowedOrigins replaces the allow-list used by the WebSocket
// upgrader. Called once at startup from cmd/server/main.go so the
// list reflects cfg.CORSOrigin rather than being hard-coded.
func SetAllowedOrigins(origins []string) {
	if len(origins) == 0 {
		return
	}
	allowedOrigins = append([]string(nil), origins...)
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[*websocket.Conn]struct{})}
}

func (h *Hub) Subscribe(companyID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[companyID] == nil {
		h.clients[companyID] = make(map[*websocket.Conn]struct{})
	}
	h.clients[companyID][conn] = struct{}{}
}

func (h *Hub) Unsubscribe(companyID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[companyID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.clients, companyID)
		}
	}
}

func (h *Hub) Broadcast(companyID string, msg any) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.clients[companyID] {
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("ws write error: %v", err)
		}
	}
}

func (h *Hub) HandleWS(c *gin.Context) {
	companyID := c.GetString("companyId")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing company ID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}
	defer conn.Close()

	h.Subscribe(companyID, conn)
	defer h.Unsubscribe(companyID, conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
