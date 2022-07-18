package go_net

import (
	"net"
	"sync"
	"time"

	"github.com/hezhis/go_log"
)

const (
	acceptDelayMax = time.Second
)

type (
	TcpServerOption func(s *TcpServer)

	TcpServer struct {
		ln     net.Listener
		lnWait sync.WaitGroup

		connWait  sync.WaitGroup
		connMutex sync.Mutex
		connSet   map[net.Conn]struct{}

		pCreateParam *CreateConnectorParam

		nMaxClientCount int
		sLocalHost      string

		NewAgent func(connector *TcpConnector) Agent
	}
)

func NewTcpServer(opts ...TcpServerOption) *TcpServer {
	s := &TcpServer{pCreateParam: &CreateConnectorParam{}}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *TcpServer) Start() {
	s.init()
	go s.logicRun()
}

func (s *TcpServer) init() {
	if s.NewAgent == nil {
		logger.Fatal("NewAgent must not be nil")
	}

	if s.pCreateParam.nHeadLength != 2 && s.pCreateParam.nHeadLength != 4 {
		logger.Fatal("message head length must be 2 or 4")
	}

	if s.pCreateParam.nMaxMsgLength <= 0 {
		s.pCreateParam.nMaxMsgLength = 4096
		logger.Warn("invalid MaxMsgLen, reset to %v", s.pCreateParam.nMaxMsgLength)
	}

	ln, err := net.Listen("tcp", s.sLocalHost)
	if nil != err {
		logger.Fatal("tcp server listen error!", err)
	}

	if s.nMaxClientCount <= 0 {
		s.nMaxClientCount = 100
		logger.Info("invalid nMaxClientCount, reset to %v", s.nMaxClientCount)
	}

	if s.pCreateParam.nWriteBuffCap <= 0 {
		s.pCreateParam.nWriteBuffCap = 100
		logger.Info("invalid nWriteBuffCap, reset to %v", s.pCreateParam.nWriteBuffCap)
	}

	s.ln = ln
	s.connSet = make(map[net.Conn]struct{})
}

func (s *TcpServer) logicRun() {
	s.lnWait.Add(1)
	defer s.lnWait.Done()

	var delay time.Duration
	for {
		conn, err := s.ln.Accept()
		if nil != err {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if delay == 0 {
					delay = 5 * time.Millisecond
				} else {
					delay *= 2
				}
				if delay > acceptDelayMax {
					delay = acceptDelayMax
				}
				time.Sleep(delay)
				continue
			}
			return
		}
		delay = 0

		s.connMutex.Lock()
		if len(s.connSet) >= s.nMaxClientCount {
			logger.Error("too many connections")
			s.connMutex.Unlock()
			conn.Close()
			continue
		}
		s.connSet[conn] = struct{}{}
		s.connMutex.Unlock()

		s.connWait.Add(1)

		connector := newTcpConnector(conn, s.pCreateParam)
		agent := s.NewAgent(connector)
		go func() {
			agent.LogicRun()

			connector.Close()

			s.connMutex.Lock()
			delete(s.connSet, conn)
			s.connMutex.Unlock()

			agent.OnClose()

			s.connWait.Done()
		}()
	}
}

func (s *TcpServer) Close() {
	s.ln.Close()
	s.lnWait.Wait()

	s.connMutex.Lock()
	for conn := range s.connSet {
		conn.Close()
	}
	s.connSet = nil
	s.connMutex.Unlock()

	s.connWait.Wait()
}
