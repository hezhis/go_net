package go_net

import (
	"log"
	"net"
	"sync"
	"time"
)

type (
	TcpClient struct {
		param *CreateConnectorParam

		conn            net.Conn
		locker          *Locker
		wg              sync.WaitGroup
		closed          bool
		connectInterval time.Duration
		autoReconnect   bool
		remoteAddr      string

		NewAgent func(connector *TcpConnector) Agent
	}
	TcpClientOption func(c *TcpClient)
)

func NewTcpClient(opts ...TcpClientOption) *TcpClient {
	client := &TcpClient{}
	client.locker = NewLocker()
	for _, opt := range opts {
		opt(client)
	}
	return client
}

func (client *TcpClient) Start() {
	client.init()

	client.wg.Add(1)
	go client.connect()
}

func (client *TcpClient) init() {
	client.locker.Lock()
	defer client.locker.Unlock()

	if nil == client.NewAgent {
		log.Fatal("NewAgent must not be nil")
	}

	if client.connectInterval <= 0 {
		client.connectInterval = 3 * time.Second
	}

}

func (client *TcpClient) dial() net.Conn {
	for {
		conn, err := net.Dial("tcp", client.remoteAddr)
		if err == nil || client.closed {
			return conn
		}

		log.Printf("connect to %v error: %v", client.remoteAddr, err)
		time.Sleep(client.connectInterval)
		continue
	}
}

func (client *TcpClient) connect() {
	defer client.wg.Done()

reconnect:
	conn := client.dial()
	if conn == nil {
		return
	}

	client.locker.Lock()
	if client.closed {
		client.locker.Unlock()
		conn.Close()
		return
	}
	client.conn = conn
	client.locker.Unlock()

	tcpConn := newTcpConnector(conn, client.param)
	agent := client.NewAgent(tcpConn)
	agent.LogicRun()

	// cleanup
	tcpConn.Close()
	client.locker.Lock()
	client.conn = nil
	client.locker.Unlock()
	agent.OnClose()

	if client.autoReconnect {
		time.Sleep(client.connectInterval)
		goto reconnect
	}
}

func (client *TcpClient) Close() {
	client.locker.Lock()
	client.closed = true

	client.conn.Close()

	client.locker.Unlock()

	client.wg.Wait()
}
