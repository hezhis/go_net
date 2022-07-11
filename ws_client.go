package go_net

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type (
	WSClient struct {
		pParam  *CreateConnectorParam
		pConn   *websocket.Conn
		pLocker *Locker

		bAutoReconnect    bool
		bClosed           bool
		sRemoteAddr       string
		nConnectInterval  time.Duration
		nHandshakeTimeout time.Duration

		NewAgent func(*WSConnector) Agent

		dialer    websocket.Dialer
		uConnWait sync.WaitGroup
	}
	WSClientOption func(c *WSClient)
)

func NewWSClient(opts ...WSClientOption) *WSClient {
	client := &WSClient{pParam: &CreateConnectorParam{}}
	client.pLocker = NewLocker()
	for _, opt := range opts {
		opt(client)
	}
	return client
}

func (c *WSClient) Start() {
	c.init()

	c.uConnWait.Add(1)
	go c.connect()
}

func (c *WSClient) init() {
	c.pLocker.Lock()
	defer c.pLocker.Unlock()

	if c.nConnectInterval <= 0 {
		c.nConnectInterval = 3 * time.Second
		log.Printf("invalid nConnectInterval, reset to %v", c.nConnectInterval)
	}
	if c.pParam.nWriteBuffCap <= 0 {
		c.pParam.nWriteBuffCap = 100
		log.Printf("invalid nWriteBuffCap, reset to %v", c.pParam.nWriteBuffCap)
	}
	if c.pParam.nMaxMsgLength <= 0 {
		c.pParam.nMaxMsgLength = 4096
		log.Printf("invalid MaxMsgLen, reset to %v", c.pParam.nMaxMsgLength)
	}
	if c.nHandshakeTimeout <= 0 {
		c.nHandshakeTimeout = 10 * time.Second
		log.Printf("invalid nHandshakeTimeout, reset to %v", c.nHandshakeTimeout)
	}
	if c.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}

	c.bClosed = false
	c.dialer = websocket.Dialer{
		HandshakeTimeout: c.nHandshakeTimeout,
	}
}

func (c *WSClient) dial() *websocket.Conn {
	for {
		conn, _, err := c.dialer.Dial(c.sRemoteAddr, nil)
		if err == nil || c.bClosed {
			return conn
		}

		log.Printf("connect to %v error: %v", c.sRemoteAddr, err)
		time.Sleep(c.nConnectInterval)
		continue
	}
}

func (c *WSClient) connect() {
	defer c.uConnWait.Done()

reconnect:
	conn := c.dial()
	if conn == nil {
		return
	}
	conn.SetReadLimit(int64(c.pParam.nMaxMsgLength))

	c.pLocker.Lock()
	if c.bClosed {
		c.pLocker.Unlock()
		conn.Close()
		return
	}
	c.pConn = conn
	c.pLocker.Unlock()

	wsConn := newWSConnector(conn, c.pParam)
	agent := c.NewAgent(wsConn)
	agent.LogicRun()

	// cleanup
	wsConn.Close()
	c.pLocker.Lock()
	c.pConn = nil
	c.pLocker.Unlock()
	agent.OnClose()

	if c.bAutoReconnect {
		time.Sleep(c.nConnectInterval)
		goto reconnect
	}
}

func (c *WSClient) Close() {
	c.pLocker.Lock()
	if !c.bClosed {
		c.bClosed = true
		if nil != c.pConn {
			c.pConn.Close()
			c.pConn = nil
		}
	}

	c.pLocker.Unlock()
	c.uConnWait.Wait()
}
