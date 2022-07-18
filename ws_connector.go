package go_net

import (
	"errors"
	"net"

	"github.com/gorilla/websocket"
	"github.com/hezhis/go_log"
)

type WSConnector struct {
	pConn   *websocket.Conn
	pLocker *Locker

	bClosed bool

	nMaxMsgLength  uint32
	cWriteBuffChan chan []byte
}

func newWSConnector(conn *websocket.Conn, param *CreateConnectorParam) *WSConnector {
	c := &WSConnector{}
	c.pLocker = NewLocker()
	c.pConn = conn

	if 0 == param.nWriteBuffCap {
		param.nWriteBuffCap = 1024
	}

	c.nMaxMsgLength = param.nMaxMsgLength
	c.cWriteBuffChan = make(chan []byte, param.nWriteBuffCap)

	go c.startWriter(conn)

	return c
}

func (c *WSConnector) startWriter(conn *websocket.Conn) {
	for b := range c.cWriteBuffChan {
		if b == nil {
			break
		}

		if err := conn.WriteMessage(websocket.BinaryMessage, b); nil != err {
			logger.Error("%v", err)
			break
		}
	}

	conn.Close()

	c.pLocker.Lock()
	c.bClosed = true
	c.pLocker.Unlock()
}

func (c *WSConnector) doDestroy() {
	c.pConn.UnderlyingConn().(*net.TCPConn).SetLinger(0)
	c.pConn.Close()

	if !c.bClosed {
		close(c.cWriteBuffChan)
		c.bClosed = true
	}
}

func (c *WSConnector) Destroy() {
	c.pLocker.Lock()
	defer c.pLocker.Unlock()

	c.doDestroy()
}

func (c *WSConnector) Close() {
	c.pLocker.Lock()
	defer c.pLocker.Unlock()
	if c.bClosed {
		return
	}

	c.doWrite(nil)
	c.bClosed = true
}

func (c *WSConnector) doWrite(b []byte) {
	if len(c.cWriteBuffChan) == cap(c.cWriteBuffChan) {
		logger.Error("close conn: channel full")
		c.doDestroy()
		return
	}

	c.cWriteBuffChan <- b
}

func (c *WSConnector) LocalAddr() net.Addr {
	return c.pConn.LocalAddr()
}

func (c *WSConnector) RemoteAddr() net.Addr {
	return c.pConn.RemoteAddr()
}

func (c *WSConnector) ReadMsg() ([]byte, error) {
	_, b, err := c.pConn.ReadMessage()
	return b, err
}

func (c *WSConnector) WriteMsg(args ...[]byte) error {
	c.pLocker.Lock()
	defer c.pLocker.Unlock()
	if c.bClosed {
		return nil
	}

	// get len
	var msgLen uint32
	for i := 0; i < len(args); i++ {
		msgLen += uint32(len(args[i]))
	}

	// check len
	if msgLen > c.nMaxMsgLength {
		return errors.New("message too long")
	} else if msgLen < 1 {
		return errors.New("message too short")
	}

	// don't copy
	if len(args) == 1 {
		c.doWrite(args[0])
		return nil
	}

	// merge the args
	msg := make([]byte, msgLen)
	l := 0
	for i := 0; i < len(args); i++ {
		copy(msg[l:], args[i])
		l += len(args[i])
	}

	c.doWrite(msg)

	return nil
}
