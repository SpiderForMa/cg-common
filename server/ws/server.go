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
	clients  map[string]map[string]*websocket.Conn
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
		clients: make(map[string]map[string]*websocket.Conn),
	}
}

// HandleConnection 处理 WebSocket 连接
func (ws *WebSocketServer) HandleConnection(w http.ResponseWriter, r *http.Request) {

	// 从查询参数获取用户 ID
	gameID := r.URL.Query().Get("gameId")
	userID := r.URL.Query().Get("userId")
	if gameID == "" || userID == "" {
		http.Error(w, "gameId and userId are required", http.StatusBadRequest)
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
	if ws.clients[gameID] == nil {
		ws.clients[gameID] = make(map[string]*websocket.Conn)
	}
	ws.clients[gameID][userID] = conn
	ws.mu.Unlock()

	defer func() {
		ws.handler.OnClose(conn)
		ws.mu.Lock()
		delete(ws.clients[gameID], userID) // 从游戏的连接列表中移除
		ws.mu.Unlock()
		conn.Close()
	}()

	// 处理消息
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
func (ws *WebSocketServer) Broadcast(gameID string, message []byte) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	for _, conn := range ws.clients[gameID] {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			ws.handler.OnError(err)
			conn.Close()
		}
	}
}

// SendToUser 向指定用户发送消息
func (ws *WebSocketServer) SendToUser(gameID string, userID string, message []byte) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	conn, exists := ws.clients[gameID][userID]
	if exists {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			ws.handler.OnError(err)
			conn.Close()
			delete(ws.clients[gameID], userID)
		}
	} else {
		fmt.Printf("用户 %s 不在线\n", userID)
	}
}

// SendToUsers 向指定用户列表发送消息
func (ws *WebSocketServer) SendToUsers(gameID string, userIDs []string, message []byte) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	for _, userID := range userIDs {
		conn, exists := ws.clients[gameID][userID]
		if exists {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				ws.handler.OnError(err)
				conn.Close()
				delete(ws.clients[gameID], userID)
			}
		} else {
			fmt.Printf("用户 %s 不在线\n", userID)
		}
	}
}
