package cde

import (
	"log"
	"strconv"

	"github.com/gorilla/websocket"
)

// Server handle all connections through websocket, distribute messages to sub-handlers
// 解析消息，根据消息类型执行内部操作或转发消息
// 内部消息有软件管理，数据拉取、项目准备等
// 一个基本的流程是：
// 从 client 来一个新的连接，client会带上项目的初始化信息（仓库地址、软件依赖、转发规则、nfs挂载等）
// server 根据 client 传入的模版，启动 ssh 转发（通过 websocket，要考虑同时打开多个项目的情况），拉取目标仓库，挂载通用 NFS，
type Server struct {
	connectionManager ConnectionManager
	messageParser     *MessageParser

	// 消息总线
	msgChan chan Message
}

// The `OnConnection` method in the `Server` struct is responsible for handling messages received and
// sent over a WebSocket connection. Here's a breakdown of what the method does:
// This code snippet defines a method `oNcONNection` on the `Server` struct in the `cde` package. This
// method is responsible for handling messages sent and received over a websocket connection.
// websocket连接上之后，在这里处理消息收发
func (s *Server) OnConnection(conn *websocket.Conn) {
	s.messageParser = &MessageParser{conn, s.msgChan, false}

	// receive and parse messages from client
	go s.messageParser.parse()

	// manage connections to the host and forward messages
	for msg := range s.msgChan {
		switch msg.msgType {
		case MSG_INIT:
			// system init
			log.Println("MSG_INIT client send init message: ", string(msg.content))
		case MSG_FORWARD:
			// send the message to ConnectionManager or client
			if msg.fromServer {
				conn.WriteMessage(websocket.BinaryMessage, msg.Raw())
			} else {
				// send message to Connection Manager
				if err := s.connectionManager.WriteMessage(msg); err != nil {
					log.Fatal("error write to local port: ", err, msg.content)
				}

			}
		case MSG_CONN_CREATE:
			// create a new port group
			log.Println("MSG_CONN_CREATE connection create: ", msg.channel, string(msg.content))
			port, _ := strconv.Atoi(string(msg.content))
			s.connectionManager.addForwardConnection(msg.channel, port)
		}
	}
}

func NewServer() (server Server) {
	msgChan := make(chan Message, 10)
	server = Server{
		msgChan:           msgChan,
		connectionManager: NewConnectionManager(msgChan),
		// MessageParser is not ready in this stage
	}
	return
}
