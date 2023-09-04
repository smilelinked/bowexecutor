/*
 * @Author: fzf404
 * @Date: 2021-09-21 10:16:53
 * @LastEditTime: 2021-09-22 16:47:32
 * @Description: Socket服务
 */
package service

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"k8s.io/klog/v2"
)

var manager *socketManager

const (
	writeWait = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 跨域
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type socketManager struct {
	clients    map[*socketClient]bool
	register   chan *socketClient
	unregister chan *socketClient
	receive    chan map[*socketClient][]byte
	broadcast  chan []byte
}

/**
 * @description: 创建Socket管理器
 */
func newManager() *socketManager {
	return &socketManager{
		clients:    make(map[*socketClient]bool),
		register:   make(chan *socketClient),
		unregister: make(chan *socketClient),
		receive:    make(chan map[*socketClient][]byte),
		broadcast:  make(chan []byte), // 广播
	}
}

func (m *socketManager) run() {
	for {
		select {
		// 连接加入
		case client := <-m.register:
			// 设置clinet为true
			m.clients[client] = true
		// 连接关闭
		case client := <-m.unregister:
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				close(client.send)
			}
		// 发送message
		case message := <-m.broadcast:
			for client := range m.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(m.clients, client)
				}
			}
		}
	}
}

type socketClient struct {
	name    uint64
	manager *socketManager
	conn    *websocket.Conn
	send    chan []byte
}

func (c *socketClient) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("read error, but not really an error, this connection has been closed, so we can't read any more.: %v", err)
			break
		}
		c.manager.receive <- map[*socketClient][]byte{
			c: message,
		}
	}
}

func (c *socketClient) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		message, ok := <-c.send
		if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil || !ok {
			err = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			if err != nil {
				klog.Info("we will close this connection.")
			}
			return
		}

		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		if _, err = w.Write(message); err != nil {
			return
		}

		// Add queued chat messages to the current websocket message.
		n := len(c.send)
		for i := 0; i < n; i++ {
			if _, err = w.Write(<-c.send); err != nil {
				return
			}
		}

		if err := w.Close(); err != nil {
			return
		}
	}
}

func SendClientSocket(name uint64, message string) {
	for k := range manager.clients {
		if k.name == name {
			k.send <- []byte(message)
		}
	}
}

func SendAllSocket(message string) {
	manager.broadcast <- []byte(message)
}

func SocketServer(w http.ResponseWriter, r *http.Request, id uint64) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &socketClient{name: id, manager: manager, conn: conn, send: make(chan []byte, 256)}
	client.manager.register <- client
	go client.writePump()
	go client.readPump()
}

func SocketInit() {
	manager = newManager()
	go manager.run()
}

func CloseClientSocket(name uint64) {
	for k := range manager.clients {
		if k.name == name {
			manager.unregister <- k
			k.conn.Close()
		}
	}
}
