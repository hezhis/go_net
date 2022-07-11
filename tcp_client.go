package go_net

import (
	"log"
	"net"
	"sync"
	"time"
)

type (
	TcpClient struct {
		pCreateParam *CreateConnectorParam
		pLocker      *Locker

		bClosed          bool
		bAutoReconnect   bool
		nConnectInterval time.Duration
		sRemoteAddr      string

		conn net.Conn
		wg   sync.WaitGroup

		NewAgent func(connector *TcpConnector) Agent
	}
	TcpClientOption func(c *TcpClient)
)

func NewTcpClient(opts ...TcpClientOption) *TcpClient {
	client := &TcpClient{pCreateParam: &CreateConnectorParam{}}
	client.pLocker = NewLocker()
	for _, opt := range opts {
		opt(client)
	}
	return client
}

func (c *TcpClient) Start() {
	c.init()

	c.wg.Add(1)
	go c.connect()
}

func (c *TcpClient) init() {
	c.pLocker.Lock()
	defer c.pLocker.Unlock()

	if nil == c.NewAgent {
		log.Fatal("NewAgent must not be nil")
	}

	if c.pCreateParam.nMaxMsgLength <= 0 {
		c.pCreateParam.nMaxMsgLength = 4096
		log.Printf("invalid MaxMsgLen, reset to %v", c.pCreateParam.nMaxMsgLength)
	}

	if c.nConnectInterval <= 0 {
		c.nConnectInterval = 3 * time.Second
	}

}

func (c *TcpClient) dial() net.Conn {
	for {
		conn, err := net.Dial("tcp", c.sRemoteAddr)
		if err == nil || c.bClosed {
			return conn
		}

		log.Printf("connect to %v error: %v", c.sRemoteAddr, err)
		time.Sleep(c.nConnectInterval)
		continue
	}
}

func (c *TcpClient) connect() {
	defer c.wg.Done()

reconnect:
	conn := c.dial()
	if conn == nil {
		return
	}

	c.pLocker.Lock()
	if c.bClosed {
		c.pLocker.Unlock()
		conn.Close()
		return
	}
	c.conn = conn
	c.pLocker.Unlock()

	tcpConn := newTcpConnector(conn, c.pCreateParam)
	agent := c.NewAgent(tcpConn)
	agent.LogicRun()

	// cleanup
	tcpConn.Close()
	c.pLocker.Lock()
	c.conn = nil
	c.pLocker.Unlock()
	agent.OnClose()

	if c.bAutoReconnect {
		time.Sleep(c.nConnectInterval)
		goto reconnect
	}
}

func (c *TcpClient) Close() {
	c.pLocker.Lock()

	if !c.bClosed {
		c.bClosed = true
		if nil != c.conn {
			c.conn.Close()
			c.conn = nil
		}
	}
	c.pLocker.Unlock()

	c.wg.Wait()
}
