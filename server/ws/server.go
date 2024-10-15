package ws

import (
	"github.com/gorilla/websocket"
	"net/http"
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
	}
}

// HandleConnection 处理 WebSocket 连接
func (ws *WebSocketServer) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.handler.OnError(err)
		return
	}
	ws.handler.OnOpen(conn)

	defer func() {
		ws.handler.OnClose(conn)
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
