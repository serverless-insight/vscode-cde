package cde

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"math/rand"
)

type connection struct {
	channel   int
	port      int
	connected bool
	conn      net.Conn
	msgChan   chan Message
	hostChan  chan Message
}

func (c *connection) establish() (err error) {
	c.conn, err = net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", c.port))
	log.Println("dial local port: ", c.port)
	if err != nil {
		return
	}

	c.connected = true
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := c.conn.Read(buffer)
			if err != nil {
				if err != io.EOF {
					log.Println("Failed to read from localport:", err)
					return
				}
			}
			content := make([]byte, n)
			copy(content, buffer[:n])
			msg := Message{
				msgType:    MSG_FORWARD,
				channel:    c.channel,
				fromServer: true,
				content:    content,
			}
			c.msgChan <- msg
			if err == io.EOF {
				c.connected = false
				break
			}
		}
	}()
	return
}

func (c connection) serve() (err error) {
	buffer := make([]byte, 1024)
	for {
		n, err := c.conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Println("Connection closed")
				c.msgChan <- Message{
					msgType: MSG_CONN_CLOSE,
					channel: c.channel,
					content: []byte(strconv.Itoa(c.port)),
				}
				break
			}
			if err != io.EOF {
				log.Println("Failed to read from the client:", err)
				// connection closed, just return and close current connection
				// panic(err)
				return nil
			}
		}
		content := make([]byte, n)
		copy(content, buffer[:n])

		msg := Message{
			msgType: MSG_FORWARD,
			channel: c.channel,
			content: content,
		}
		c.msgChan <- msg
	}
	return
}

func (c *connection) send(msg Message) (err error) {
	// 如果连接还没有建立，就尝试建立连接，如果仍然失败，返回错误
	// 只有 ForwardConnection 需要建立连接
	if !c.connected {
		if err = c.establish(); err != nil {
			return
		}
		c.connected = true
	}

	_, err = c.conn.Write(msg.content)
	return
}

func newForwardConnection(channel, port int, msgChan chan Message) (c connection) {
	c = connection{
		channel:   channel,
		port:      port,
		connected: false,
		msgChan:   msgChan,
		hostChan:  make(chan Message),
	}
	if err := c.establish(); err != nil {
		// TODO: 需要 client 和 server 对齐连接失败的处理，可以用特殊的 msg type 通知 client
		msgChan <- NewErrorMessage(channel, err)
		log.Fatal(err)
		return
	}

	return
}

func newListenerConnection(conn net.Conn, port int, msgChan chan Message, callback OnListenerConnectionCallback) (c connection) {
	c = connection{
		conn:      conn,
		port:      port,
		connected: true, // 在创建 ServeConnection 的时候，已经有 net.Conn 对象了
		msgChan:   msgChan,
		hostChan:  make(chan Message),
	}
	callback(&c) // 一般是设置 channel
	go c.serve()
	return
}

const (
	PortGroupTypeListener = iota
	PortGroupTypeForward  = iota
)

// PortGroup 负责维护某一个端口下的所有连接
// 对服务端来说，只需要维护 connection 中的 conn即可，PortGroup没有太大意义
// 对客户端，因为需要作为 tcp 服务器，所以需要维护 listener，
// 所有在这个端口的连接都需要这个 listener 来接收
type PortGroup struct {
	groupType   int
	port        int
	connections map[int]*connection // map[channel]connection

	// only for PortGroupTypeListener
	listener net.Listener
}

func (pg *PortGroup) Serve(channel int, msgChan chan Message, callback OnListenerConnectionCallback) {
	for {
		// 接受新的连接
		conn, err := pg.listener.Accept()
		log.Println("Accept new connection")
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// channel 需要在 callback 中设置，此时先置为 0
		connection := newListenerConnection(conn, pg.port, msgChan, callback)
		pg.connections[channel] = &connection
	}
}

type ConnectionManager struct {
	msgChan    chan Message
	portGroups map[int]PortGroup

	connections []*connection
}

type OnListenerConnectionCallback func(c *connection)

func (cm *ConnectionManager) CloseConnection(channel int) {
	conn := cm.connections[channel]
	if conn.connected {
		conn.conn.Close()
	}
	cm.connections[channel] = nil
}

func (cm *ConnectionManager) FetchForwardGroup(port int) (pg PortGroup) {
	pg, exists := cm.portGroups[port]
	if !exists {
		pg = cm.addForwardGroup(port)
	}
	return
}

func (cm *ConnectionManager) addForwardGroup(port int) (pg PortGroup) {
	pg = PortGroup{
		groupType:   PortGroupTypeForward,
		port:        port,
		connections: make(map[int]*connection),
	}
	cm.portGroups[port] = pg
	return
}

func (cm *ConnectionManager) addForwardConnection(channel, port int) (c connection) {
	c = newForwardConnection(channel, port, cm.msgChan)
	pg := cm.FetchForwardGroup(port)
	pg.connections[channel] = &c

	cm.connections[channel] = &c
	return
}

// FetchListnerGroup 获取或初始化一组监听端口
func (cm *ConnectionManager) FetchListenerGroup(port int) (pg PortGroup, err error) {
	pg, exists := cm.portGroups[port]
	log.Println("fetch listener group: ", port)
	if !exists {
		pg, err = cm.addListenerGroup(port, cm.availableLocalPort())
	}
	return
}

// addListenerGroup 创建一个新的监听端组，监听一个随机端口(注意这个是给客户端用的)
func (cm *ConnectionManager) addListenerGroup(port, localPort int) (pg PortGroup, err error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", localPort))
	log.Printf("Listening on local port: %d, remote port: %d", localPort, port)
	if err != nil {
		return
	}

	pg = PortGroup{
		groupType:   PortGroupTypeListener,
		port:        localPort,
		connections: make(map[int]*connection),
		listener:    listener,
	}
	cm.portGroups[port] = pg

	go pg.Serve(cm.availableChannel(), cm.msgChan, func(c *connection) {
		c.channel = cm.availableChannel()

		cm.connections[c.channel] = c

		log.Println("connection create: ", c.channel, localPort, port)
		// connection 创建完成后，要通知 server
		cm.msgChan <- Message{
			msgType: MSG_CONN_CREATE,
			channel: c.channel,
			content: []byte(strconv.Itoa(port)),
		}
	})
	return
}

func (cm *ConnectionManager) availableChannel() (channel int) {
	// TODO: 需要锁来保证不会有重复分配的 channel
	bitset := make([]interface{}, 256)
	for _, conn := range cm.connections {
		if conn != nil {
			bitset[conn.channel] = struct{}{}
		}
	}
	for i, v := range bitset {
		if v == nil {
			return i
		}
	}
	panic("channel exceed limit")
}

func (cm *ConnectionManager) availableLocalPort() (port int) {
	for {
		// 生成一个 9000 到 65535 之间的随机端口
		port = rand.Intn(65535-9000) + 9000
		addr := fmt.Sprintf(":%d", port)

		// 尝试监听该端口
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			// 如果端口被占用，继续尝试下一个随机端口
			continue
		}
		// 端口可用，关闭监听器并返回端口
		defer listener.Close()
		return port
	}
}

func (cm *ConnectionManager) WriteMessage(msg Message) (err error) {
	err = cm.connections[msg.channel].send(msg)
	return
}

func NewConnectionManager(msgChan chan Message) (cm ConnectionManager) {
	cm = ConnectionManager{
		msgChan:     msgChan,
		connections: make([]*connection, 255), // 最多有 255 个连接，先全部置为 nil
		portGroups:  make(map[int]PortGroup),
	}
	return
}
