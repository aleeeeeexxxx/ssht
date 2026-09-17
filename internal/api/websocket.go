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

type LogBroadcaster struct {
	clients map[*websocket.Conn]bool
	mu      sync.RWMutex
}

var logBroadcaster = &LogBroadcaster{
	clients: make(map[*websocket.Conn]bool),
}

func (lb *LogBroadcaster) AddClient(conn *websocket.Conn) {
	lb.mu.Lock()
	lb.clients[conn] = true
	lb.mu.Unlock()
}

func (lb *LogBroadcaster) RemoveClient(conn *websocket.Conn) {
	lb.mu.Lock()
	delete(lb.clients, conn)
	lb.mu.Unlock()
}

func (lb *LogBroadcaster) Broadcast(entry LogEntry) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	for conn := range lb.clients {
		conn.WriteMessage(websocket.TextMessage, data)
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

	logBroadcaster.AddClient(conn)
	defer logBroadcaster.RemoveClient(conn)

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
