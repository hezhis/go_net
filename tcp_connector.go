package go_net

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"

	"github.com/hezhis/go_log"
)

type (
	TcpConnector struct {
		conn net.Conn

		bClosed       bool
		bLittleEndian bool
		nHeadLength   int
		nMaxMsgLength uint32

		pLocker *Locker

		bHeader        []byte
		cWriteBuffChan chan []byte
	}

	CreateConnectorParam struct {
		nHeadLength   int
		nWriteBuffCap int
		nMaxMsgLength uint32
		bLittleEndian bool
	}
)

func newTcpConnector(conn net.Conn, param *CreateConnectorParam) *TcpConnector {
	c := &TcpConnector{}
	c.pLocker = NewLocker()
	c.conn = conn

	if 0 == param.nHeadLength {
		param.nHeadLength = 2
	}
	if 0 == param.nWriteBuffCap {
		param.nWriteBuffCap = 1024
	}

	var max uint32
	switch param.nHeadLength {
	case 2:
		max = math.MaxUint16
	case 4:
		max = math.MaxUint32
	}
	if max > param.nMaxMsgLength {
		max = param.nMaxMsgLength
	}

	c.nMaxMsgLength = max
	c.nHeadLength = param.nHeadLength
	c.bLittleEndian = param.bLittleEndian
	c.bHeader = make([]byte, param.nHeadLength)
	c.cWriteBuffChan = make(chan []byte, param.nWriteBuffCap)

	go c.startWriter(conn)

	return c
}

func (c *TcpConnector) startWriter(conn net.Conn) {
	for b := range c.cWriteBuffChan {
		if b == nil {
			break
		}

		if _, err := conn.Write(b); nil != err {
			logger.Error("write data error! %v", err)
			break
		}
	}

	conn.Close()

	c.pLocker.Lock()
	c.bClosed = true
	c.pLocker.Unlock()
}

func (c *TcpConnector) Write(b []byte) {
	if nil == b {
		return
	}
	c.pLocker.Lock()
	defer c.pLocker.Unlock()

	if c.bClosed {
		return
	}

	c.doWrite(b)
}

func (c *TcpConnector) doWrite(b []byte) {
	if len(c.cWriteBuffChan) == cap(c.cWriteBuffChan) {
		logger.Error("close conn: channel full")
		c.doDestroy()
		return
	}

	c.cWriteBuffChan <- b
}

func (c *TcpConnector) Close() {
	c.pLocker.Lock()
	defer c.pLocker.Unlock()
	if c.bClosed {
		return
	}

	c.doWrite(nil)
	c.bClosed = true
}

func (c *TcpConnector) Destroy() {
	c.pLocker.Lock()
	defer c.pLocker.Unlock()

	c.doDestroy()
}

func (c *TcpConnector) doDestroy() {
	c.conn.(*net.TCPConn).SetLinger(0)
	c.conn.Close()

	if !c.bClosed {
		close(c.cWriteBuffChan)
		c.bClosed = true
	}
}

func (c *TcpConnector) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *TcpConnector) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *TcpConnector) ReadMsg() ([]byte, error) {
	// read len
	if _, err := io.ReadFull(c.conn, c.bHeader); err != nil {
		return nil, err
	}

	var msgLen uint32
	switch c.nHeadLength {
	case 2:
		if c.bLittleEndian {
			msgLen = uint32(binary.LittleEndian.Uint16(c.bHeader))
		} else {
			msgLen = uint32(binary.BigEndian.Uint16(c.bHeader))
		}
	case 4:
		if c.bLittleEndian {
			msgLen = binary.LittleEndian.Uint32(c.bHeader)
		} else {
			msgLen = binary.BigEndian.Uint32(c.bHeader)
		}
	}

	// check len
	if msgLen > c.nMaxMsgLength {
		return nil, errors.New("message too long")
	}

	// data
	msgData := make([]byte, msgLen)
	if _, err := io.ReadFull(c.conn, msgData); err != nil {
		return nil, err
	}

	return msgData, nil
}

func (c *TcpConnector) WriteMsg(args ...[]byte) error {
	// get len
	var msgLen uint32
	for i := 0; i < len(args); i++ {
		msgLen += uint32(len(args[i]))
	}

	// check len
	if msgLen > c.nMaxMsgLength {
		return errors.New("message too long")
	}

	msg := make([]byte, uint32(c.nHeadLength)+msgLen)

	// write len
	switch c.nHeadLength {
	case 2:
		if c.bLittleEndian {
			binary.LittleEndian.PutUint16(msg, uint16(msgLen))
		} else {
			binary.BigEndian.PutUint16(msg, uint16(msgLen))
		}
	case 4:
		if c.bLittleEndian {
			binary.LittleEndian.PutUint32(msg, msgLen)
		} else {
			binary.BigEndian.PutUint32(msg, msgLen)
		}
	}

	// write data
	l := c.nHeadLength
	for i := 0; i < len(args); i++ {
		copy(msg[l:], args[i])
		l += len(args[i])
	}

	c.Write(msg)

	return nil
}
