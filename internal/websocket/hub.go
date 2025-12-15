package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有来源的连接，生产环境应配置为前端域名
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client 是一个websocket连接的封装
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// Hub 维护了所有活跃的客户端，并向他们广播消息
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run 启动Hub的事件循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Println("New client registered")
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Println("Client unregistered")
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// 如果发送缓冲区已满，则关闭连接
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast 广播消息给所有客户端
func (h *Hub) Broadcast(message interface{}) {
	jsonMsg, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling broadcast message: %v", err)
		return
	}
	h.broadcast <- jsonMsg
}

// ServeWs 处理websocket请求
func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request, onConnect func() interface{}) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: h, conn: conn, send: make(chan []byte, 256)}
	h.register <- client

	// 当新客户端连接时，立即发送当前状态
	initialState := onConnect()
	if initialState != nil {
		jsonState, err := json.Marshal(initialState)
		if err == nil {
			client.send <- jsonState
		}
	}

	go client.writePump()
	// 我们主要通过HTTP API控制，所以readPump可以很简单
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	// 简单地丢弃所有收到的消息，只用于检测连接是否断开
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}
