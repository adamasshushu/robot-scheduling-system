package service

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// WSMessage WebSocket 推送消息
type WSMessage struct {
	Type string      `json:"type"` // telemetry | alert | task_status | robot_status
	Data interface{} `json:"data"`
}

// WSClient 单个 WebSocket 连接
type WSClient struct {
	Conn *websocket.Conn
	Send chan []byte
	Hub  *WSHub
}

// WSHub WebSocket 管理中心
type WSHub struct {
	mu      sync.RWMutex
	Clients map[*WSClient]bool
}

var Hub = &WSHub{
	Clients: make(map[*WSClient]bool),
}

// Register 注册新客户端
func (h *WSHub) Register(client *WSClient) {
	h.mu.Lock()
	h.Clients[client] = true
	h.mu.Unlock()
	log.Printf("🔌 WS client connected — total: %d", len(h.Clients))
}

// Unregister 注销客户端
func (h *WSHub) Unregister(client *WSClient) {
	h.mu.Lock()
	if _, ok := h.Clients[client]; ok {
		delete(h.Clients, client)
		close(client.Send)
	}
	h.mu.Unlock()
	log.Printf("🔌 WS client disconnected — total: %d", len(h.Clients))
}

// Broadcast 广播消息给所有客户端
func (h *WSHub) Broadcast(msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("❌ WS marshal error: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.Clients {
		select {
		case client.Send <- data:
		default:
			// 客户端发送缓冲区满，跳过
			go h.Unregister(client)
		}
	}
}

// Count 当前连接数
func (h *WSHub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Clients)
}
