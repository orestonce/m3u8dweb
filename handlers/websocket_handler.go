// filePath: m3u8dweb/handlers/websocket_handler.go
package handlers

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许跨域，生产环境需限制
		},
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

// 仅包含需要推送的字段
type TaskProgressUpdate struct {
	ID        string `json:"id"`
	StatusBar string `json:"status_bar"`
}

// WebSocket处理器
func TaskWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket升级失败:", err)
		return
	}
	defer conn.Close()

	// 添加客户端到连接池
	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	log.Println("新客户端连接，当前连接数:", len(clients))

	// 保持连接
	for {
		_, _, err = conn.ReadMessage()
		if err != nil {
			break
		}
	}

	// 移除客户端
	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
}

// 推送任务进度更新
func BroadcastTaskThread(ch <- chan []byte ) {

	for {
		data := <- ch

		func() {
			clientsMu.Lock()
			defer clientsMu.Unlock()

			for client := range clients {
				if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
					log.Println("推送失败:", err)
					client.Close()
					delete(clients, client)
				}
			}
		}()
	}
}