package cde

import (
	"io"
	"log"

	"github.com/gorilla/websocket"
)

const (
	MSG_INIT        = iota // 客户端请求初始化
	MSG_INITTALIZED = iota // 服务端应答初始化完成
	MSG_FORWARD     = iota // 端口转发数据
	MSG_CONN_CREATE = iota // 连接创建
	MSG_CONN_CLOSE  = iota // 连接关闭
	MSG_ERROR       = iota // 错误消息
)

type Message struct {
	msgType    int
	channel    int
	fromServer bool // 默认是从 client 流向 server 的
	content    []byte
}

func NewErrorMessage(channel int, err error) (msg Message) {
	msg = Message{
		msgType:    MSG_ERROR,
		channel:    channel,
		fromServer: true,
		content:    []byte(err.Error()),
	}
	return
}

func (msg Message) Raw() (bs []byte) {
	bs = make([]byte, len(msg.content)+2)
	bs[0] = byte(msg.msgType)
	bs[1] = byte(msg.channel)
	copy(bs[2:len(msg.content)+2], msg.content)
	return
}

type MessageParser struct {
	conn       *websocket.Conn
	msgChan    chan Message
	serverSide bool
}

func (p *MessageParser) parse() (err error) {
	for {
		_, reader, err := p.conn.NextReader()
		if err != nil {
			log.Println(err)
			return err
		}
		data, err := io.ReadAll(reader)
		if err != nil {
			log.Println("read websocket message failed:", err)
			return err
		}
		if len(data) < 2 {
			log.Println("invalid message length:", len(data))
			continue
		}
		msg := Message{
			msgType:    int(data[0]),
			channel:    int(data[1]),
			fromServer: p.serverSide,
			content:    data[2:],
		}
		p.msgChan <- msg
	}
}
