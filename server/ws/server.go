package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

// WebSocketHandler 定义 WebSocket 事件处理器的接口
type WebSocketHandler interface {
	OnOpen(conn *websocket.Conn)
	OnClose(conn *websocket.Conn)
	OnError(err error)
	OnMessage(conn *websocket.Conn, msg []byte)
}

// WebSocketServer 封装 WebSocket 服务器
type WebSocketServer struct {
	upgrader websocket.Upgrader
	handler  WebSocketHandler
	clients  map[string]*websocket.Conn
	mu       sync.Mutex
}

// NewWebSocketServer 创建新的 WebSocket 服务器
func NewWebSocketServer(handler WebSocketHandler) *WebSocketServer {
	return &WebSocketServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有跨域请求
			},
		},
		handler: handler,
		clients: make(map[string]*websocket.Conn),
	}
}

// HandleConnection 处理 WebSocket 连接
func (ws *WebSocketServer) HandleConnection(w http.ResponseWriter, r *http.Request) {

	// 从查询参数获取用户 ID
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.handler.OnError(err)
		return
	}
	ws.handler.OnOpen(conn)

	// 添加到连接列表
	ws.mu.Lock()
	ws.clients[userID] = conn
	ws.mu.Unlock()

	defer func() {
		ws.handler.OnClose(conn)
		ws.mu.Lock()
		delete(ws.clients, userID)
		ws.mu.Unlock()
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			ws.handler.OnError(err)
			break
		}
		ws.handler.OnMessage(conn, msg)
	}
}

// Broadcast 向所有连接的客户端发送消息
func (ws *WebSocketServer) Broadcast(message []byte) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	for _, conn := range ws.clients {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			ws.handler.OnError(err)
			conn.Close()
		}
	}
}

// SendToUser 向指定用户发送消息
func (ws *WebSocketServer) SendToUser(userID string, message []byte) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	conn, exists := ws.clients[userID]
	if exists {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			ws.handler.OnError(err)
			conn.Close()
			delete(ws.clients, userID)
		}
	} else {
		fmt.Printf("用户 %s 不在线\n", userID)
	}
}

// SendToUsers 向指定用户列表发送消息
func (ws *WebSocketServer) SendToUsers(userIDs []string, message []byte) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	for _, userID := range userIDs {
		conn, exists := ws.clients[userID]
		if exists {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				ws.handler.OnError(err)
				conn.Close()
				delete(ws.clients, userID)
			}
		} else {
			fmt.Printf("用户 %s 不在线\n", userID)
		}
	}
}
