package attendance

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
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
