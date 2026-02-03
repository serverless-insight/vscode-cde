package cde

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	connectionManager ConnectionManager
	messageParser     *MessageParser

	msgChan chan Message
}

func (c *Client) handleMessage(conn *websocket.Conn) {
	for msg := range c.msgChan {
		switch msg.msgType {
		case MSG_INIT, MSG_CONN_CREATE:
			if err := conn.WriteMessage(websocket.BinaryMessage, msg.Raw()); err != nil {
				panic(err)
			}
		case MSG_FORWARD:
			// send the message to ConnectionManager or client
			if msg.fromServer {
				// send message to Connection Manager
				c.connectionManager.WriteMessage(msg)
			} else {
				if err := conn.WriteMessage(websocket.BinaryMessage, msg.Raw()); err != nil {
					panic(err)
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

func (c *Client) Run(conn *websocket.Conn) error {
	// 先启动消息处理
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println("error handle message: ", err)
				close(c.msgChan)
				c.messageParser.conn.Close()
				return
			}
		}()
		c.handleMessage(conn)
	}()

	// 1. send init message to server
	c.msgChan <- Message{
		msgType: MSG_INIT,
		content: []byte("init"),
	}

	// // 2. 创建一个 ssh 的 serveConnection
	if _, err := c.connectionManager.FetchListenerGroup(22); err != nil {
		log.Fatal(err)
		return err
	}

	// receive and parse messages from server
	c.messageParser = &MessageParser{conn, c.msgChan, true}
	return c.messageParser.parse()
}

func NewClient() (client Client) {
	msgChan := make(chan Message, 1024)
	client = Client{
		msgChan:           msgChan,
		connectionManager: NewConnectionManager(msgChan),
	}
	return
}
