package go_net

import (
	"log"
	"net"
	"sync"
	"time"
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

		param *CreateConnectorParam

		nMaxClientCount int

		localhost string

		NewAgent func(connector *TcpConnector) Agent
	}
)

func NewTcpServer(opts ...TcpServerOption) *TcpServer {
	s := &TcpServer{param: &CreateConnectorParam{}}
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
		log.Fatal("NewAgent must not be nil")
	}
	if s.param.nHeadLength != 2 && s.param.nHeadLength != 4 {
		log.Fatal("message head length must be 2 or 4")
	}

	ln, err := net.Listen("tcp", s.localhost)
	if nil != err {
		log.Fatalln(err)
	}

	if s.nMaxClientCount <= 0 {
		s.nMaxClientCount = 100
		log.Printf("invalid nMaxClientCount, reset to %v\n", s.nMaxClientCount)
	}

	if s.param.nWriteBuffCap <= 0 {
		s.param.nWriteBuffCap = 100
		log.Printf("invalid nWriteBuffCap, reset to %v\n", s.param.nWriteBuffCap)
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
			log.Printf("too many connections")
			s.connMutex.Unlock()
			conn.Close()
			continue
		}
		s.connSet[conn] = struct{}{}
		s.connMutex.Unlock()

		s.connWait.Add(1)

		connector := newTcpConnector(conn, s.param)
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
