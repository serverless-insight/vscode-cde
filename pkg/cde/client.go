package cde

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	connectionManager ConnectionManager
	msgChan           chan Message
}

func (c *Client) handleMessage(conn *websocket.Conn, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case msg := <-c.msgChan:
			switch msg.msgType {
			case MSG_INIT, MSG_CONN_CREATE:
				if err := conn.WriteMessage(websocket.BinaryMessage, msg.Raw()); err != nil {
					log.Println("error handle message: ", err)
					conn.Close()
					return
				}
			case MSG_FORWARD:
				// send the message to ConnectionManager or client
				if msg.fromServer {
					// send message to Connection Manager
					if err := c.connectionManager.WriteMessage(msg); err != nil {
						log.Println("error write message: ", err)
					}
				} else {
					if err := conn.WriteMessage(websocket.BinaryMessage, msg.Raw()); err != nil {
						log.Println("error handle message: ", err)
						conn.Close()
						return
					}
				}
			case MSG_CONN_CLOSE:
				// close the port group
				log.Printf("Connection closed at channel: %d, port: %s", msg.channel, string(msg.content))
				c.connectionManager.CloseConnection(msg.channel)
			case MSG_ERROR:
				// create a new port group
				log.Println("Error message received:", string(msg.content))
			}
		}
	}
}

func (c *Client) Run(conn *websocket.Conn) error {
	// // 2. 创建一个 ssh 的 serveConnection
	if _, err := c.connectionManager.FetchListenerGroup(22); err != nil {
		return err
	}

	sessionDone := make(chan struct{})
	defer close(sessionDone)
	defer c.connectionManager.CloseActiveConnections()

	// 先启动消息处理
	go c.handleMessage(conn, sessionDone)

	// 1. send init message to server
	c.msgChan <- Message{
		msgType: MSG_INIT,
		content: []byte("init"),
	}

	// receive and parse messages from server
	messageParser := &MessageParser{conn, c.msgChan, true}
	return messageParser.parse()
}

func NewClient() (client Client) {
	msgChan := make(chan Message, 1024)
	client = Client{
		msgChan:           msgChan,
		connectionManager: NewConnectionManager(msgChan),
	}
	return
}
