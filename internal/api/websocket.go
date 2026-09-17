package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

type wsClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type LogBroadcaster struct {
	clients map[*wsClient]bool
	mu      sync.RWMutex
}

var logBroadcaster = &LogBroadcaster{
	clients: make(map[*wsClient]bool),
}

func (lb *LogBroadcaster) AddClient(client *wsClient) {
	lb.mu.Lock()
	lb.clients[client] = true
	lb.mu.Unlock()
}

func (lb *LogBroadcaster) RemoveClient(client *wsClient) {
	lb.mu.Lock()
	delete(lb.clients, client)
	lb.mu.Unlock()
}

func (lb *LogBroadcaster) Broadcast(entry LogEntry) {
	lb.mu.RLock()
	clients := make([]*wsClient, 0, len(lb.clients))
	for client := range lb.clients {
		clients = append(clients, client)
	}
	lb.mu.RUnlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	for _, client := range clients {
		client.mu.Lock()
		client.conn.WriteMessage(websocket.TextMessage, data)
		client.mu.Unlock()
	}
}

// BroadcastLog sends a log entry to all connected WebSocket clients
func BroadcastLog(level, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}
	logBroadcaster.Broadcast(entry)
}

func (s *Server) handleWebSocketLogs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	client := &wsClient{conn: conn}
	logBroadcaster.AddClient(client)
	defer logBroadcaster.RemoveClient(client)

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
