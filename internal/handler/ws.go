package handler

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/adamasshushu/robot-scheduling-system/internal/service"
)

var allowedOrigins = os.Getenv("CORS_ORIGINS")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		if allowedOrigins == "" {
			return true // 开发模式允许所有来源
		}
		origin := r.Header.Get("Origin")
		for _, o := range strings.Split(allowedOrigins, ",") {
			if strings.TrimSpace(o) == origin {
				return true
			}
		}
		return false
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WSHandler WebSocket 连接处理
func WSHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ WS upgrade failed: %v", err)
		return
	}

	client := &service.WSClient{
		Conn: conn,
		Send: make(chan []byte, 64),
		Hub:  service.Hub,
	}

	service.Hub.Register(client)

	// 发送当前状态快照
	go func() {
		time.Sleep(100 * time.Millisecond)
		service.Hub.Broadcast(service.WSMessage{
			Type: "connected",
			Data: map[string]interface{}{
				"clients": service.Hub.Count(),
			},
		})
	}()

	// 写协程：从 Send channel 推送到 WebSocket
	go writePump(client)

	// 读协程：保持连接 + 处理心跳
	readPump(client)
}

func readPump(client *service.WSClient) {
	defer func() {
		client.Hub.Unregister(client)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("⚠️  WS error: %v", err)
			}
			break
		}
	}
}

func writePump(client *service.WSClient) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
