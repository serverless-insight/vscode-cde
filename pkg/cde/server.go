package cde

import (
	"log"
	"strconv"

	"github.com/gorilla/websocket"
)

type Server struct {
	connectionManager ConnectionManager
	// 消息总线
	msgChan chan Message
}

func (s *Server) OnConnection(conn *websocket.Conn) {
	messageParser := &MessageParser{conn, s.msgChan, false}
	parseErrChan := make(chan error, 1)

	// receive and parse messages from client
	go func() {
		parseErrChan <- messageParser.parse()
	}()

	// manage connections to the host and forward messages
	for {
		select {
		case err := <-parseErrChan:
			if err != nil {
				log.Println("error parse message: ", err)
			}
			s.connectionManager.CloseActiveConnections()
			return
		case msg := <-s.msgChan:
			switch msg.msgType {
			case MSG_INIT:
				// system init
				log.Println("MSG_INIT client send init message: ", string(msg.content))
			case MSG_FORWARD:
				// send the message to ConnectionManager or client
				if msg.fromServer {
					if err := conn.WriteMessage(websocket.BinaryMessage, msg.Raw()); err != nil {
						log.Println("error write websocket message: ", err)
						s.connectionManager.CloseActiveConnections()
						return
					}
				} else {
					// send message to Connection Manager
					if err := s.connectionManager.WriteMessage(msg); err != nil {
						log.Println("error write to local port: ", err, msg.content)
					}
				}
			case MSG_CONN_CREATE:
				// create a new port group
				log.Println("MSG_CONN_CREATE connection create: ", msg.channel, string(msg.content))
				port, _ := strconv.Atoi(string(msg.content))
				if _, err := s.connectionManager.addForwardConnection(msg.channel, port); err != nil {
					log.Println("error add forward connection: ", err)
				}
			}
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
