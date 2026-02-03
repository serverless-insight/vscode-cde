package cde

import (
	"log"
	"strconv"

	"github.com/gorilla/websocket"
)

type Server struct {
	connectionManager ConnectionManager
	messageParser     *MessageParser

	// 消息总线
	msgChan chan Message
}

func (s *Server) OnConnection(conn *websocket.Conn) {
	s.messageParser = &MessageParser{conn, s.msgChan, false}

	// receive and parse messages from client
	go func() {
		if err := s.messageParser.parse(); err != nil {
			log.Println("error parse message: ", err)
			close(s.msgChan)
			return
		}
	}()

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
	msgChan := make(chan Message, 1024)
	server = Server{
		msgChan:           msgChan,
		connectionManager: NewConnectionManager(msgChan),
		// MessageParser is not ready in this stage
	}
	return
}
