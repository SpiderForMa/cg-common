package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

// WebSocketHandler 定义 WebSocket 事件处理器的接口
type WebSocketHandler interface {
	OnOpen(conn *CustomConn)
	OnClose(conn *CustomConn)
	OnError(err error)
	OnMessage(conn *CustomConn, msg []byte)
}

// WebSocketServer 封装 WebSocket 服务器
type WebSocketServer struct {
	upgrader websocket.Upgrader
	handler  WebSocketHandler
	clients  map[string]map[string]*CustomConn
	mu       sync.Mutex
}

type CustomConn struct {
	Conn   *websocket.Conn
	AppID  string
	UserID string
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
		clients: make(map[string]map[string]*CustomConn),
	}
}

// HandleConnection 处理 WebSocket 连接
func (ws *WebSocketServer) HandleConnection(w http.ResponseWriter, r *http.Request) {

	appID := r.URL.Query().Get("appID")
	userID := r.URL.Query().Get("userID")
	if userID == "" || appID == "" {
		http.Error(w, "appID and userID are required", http.StatusBadRequest)
		return
	}

	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.handler.OnError(err)
		return
	}

	// 创建 CustomConn
	customConn := &CustomConn{
		Conn:   conn,
		AppID:  appID,
		UserID: userID,
	}

	ws.handler.OnOpen(customConn)

	// 添加到连接列表
	ws.mu.Lock()
	if ws.clients[appID] == nil {
		ws.clients[appID] = make(map[string]*CustomConn)
	}
	ws.clients[appID][userID] = customConn
	ws.mu.Unlock()

	defer func() {
		ws.handler.OnClose(customConn)
		ws.mu.Lock()
		delete(ws.clients[appID], userID)
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
		ws.handler.OnMessage(customConn, msg)
	}
}

// Broadcast 向所有连接的客户端发送消息
func (ws *WebSocketServer) Broadcast(gameID string, message []byte) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	for _, customConn := range ws.clients[gameID] {
		if err := customConn.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			ws.handler.OnError(err)
			customConn.Conn.Close()
		}
	}
}

// SendToUser 向指定用户发送消息
func (ws *WebSocketServer) SendToUser(gameID string, userID string, message []byte) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	customConn, exists := ws.clients[gameID][userID]
	if exists {
		if err := customConn.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			ws.handler.OnError(err)
			customConn.Conn.Close()
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
		customConn, exists := ws.clients[gameID][userID]
		if exists {
			if err := customConn.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				ws.handler.OnError(err)
				customConn.Conn.Close()
				delete(ws.clients[gameID], userID)
			}
		} else {
			fmt.Printf("用户 %s 不在线\n", userID)
		}
	}
}
