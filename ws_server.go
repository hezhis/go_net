package go_net

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hezhis/go_log"
)

type (
	WSServerOption func(s *WSServer)

	WSServer struct {
		nMaxClientCount int
		nHTTPTimeout    time.Duration
		sLocalHost      string
		sCertFile       string
		sKeyFile        string

		NewAgent func(*WSConnector) Agent

		pCreateParam *CreateConnectorParam
		pHandler     *WSHandler

		ln net.Listener
	}

	WSHandler struct {
		pCreateParam *CreateConnectorParam

		nMaxClientCount int

		pNewAgent func(*WSConnector) Agent
		upgrader  websocket.Upgrader

		connSet   map[*websocket.Conn]struct{}
		connMutex sync.Mutex
		connWait  sync.WaitGroup
	}
)

func NewWSServer(opts ...WSServerOption) *WSServer {
	s := &WSServer{pCreateParam: &CreateConnectorParam{}}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (handler *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	conn, err := handler.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("upgrade error: %v", err)
		return
	}
	conn.SetReadLimit(int64(handler.pCreateParam.nMaxMsgLength))

	handler.connWait.Add(1)
	defer handler.connWait.Done()

	handler.connMutex.Lock()
	if handler.connSet == nil {
		handler.connMutex.Unlock()
		conn.Close()
		return
	}
	if len(handler.connSet) >= handler.nMaxClientCount {
		handler.connMutex.Unlock()
		conn.Close()
		logger.Error("too many connections")
		return
	}
	handler.connSet[conn] = struct{}{}
	handler.connMutex.Unlock()

	if wsConn := newWSConnector(conn, handler.pCreateParam); nil != wsConn {
		if agent := handler.pNewAgent(wsConn); nil != agent {
			agent.LogicRun()

			// cleanup
			wsConn.Close()
			handler.connMutex.Lock()
			delete(handler.connSet, conn)
			handler.connMutex.Unlock()
			agent.OnClose()
		}
	}
}

func (s *WSServer) Start() {
	ln, err := net.Listen("tcp", s.sLocalHost)
	if err != nil {
		logger.Fatal("%v", err)
	}

	if s.nMaxClientCount <= 0 {
		s.nMaxClientCount = 100
		logger.Warn("invalid nMaxClientCount, reset to %v", s.nMaxClientCount)
	}

	if s.pCreateParam.nWriteBuffCap <= 0 {
		s.pCreateParam.nWriteBuffCap = 1024
		logger.Warn("invalid PendingWriteNum, reset to %v", s.pCreateParam.nWriteBuffCap)
	}

	if s.pCreateParam.nMaxMsgLength <= 0 {
		s.pCreateParam.nMaxMsgLength = 4096
		logger.Warn("invalid MaxMsgLen, reset to %v", s.pCreateParam.nMaxMsgLength)
	}

	if s.nHTTPTimeout <= 0 {
		s.nHTTPTimeout = 10 * time.Second
		logger.Warn("invalid nHTTPTimeout, reset to %v", s.nHTTPTimeout)
	}
	if s.NewAgent == nil {
		logger.Fatal("NewAgent must not be nil")
	}

	if s.sCertFile != "" || s.sKeyFile != "" {
		config := &tls.Config{}
		config.NextProtos = []string{"http/1.1"}

		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(s.sCertFile, s.sKeyFile)
		if err != nil {
			logger.Fatal("%v", err)
		}

		ln = tls.NewListener(ln, config)
	}

	s.ln = ln
	s.pHandler = &WSHandler{
		pCreateParam:    s.pCreateParam,
		nMaxClientCount: s.nMaxClientCount,
		pNewAgent:       s.NewAgent,
		connSet:         make(map[*websocket.Conn]struct{}),
		upgrader: websocket.Upgrader{
			HandshakeTimeout: s.nHTTPTimeout,
			CheckOrigin:      func(_ *http.Request) bool { return true },
		},
	}

	httpServer := &http.Server{
		Addr:           s.sLocalHost,
		Handler:        s.pHandler,
		ReadTimeout:    s.nHTTPTimeout,
		WriteTimeout:   s.nHTTPTimeout,
		MaxHeaderBytes: 1024,
	}

	go httpServer.Serve(ln)
}

func (s *WSServer) Close() {
	s.ln.Close()

	s.pHandler.connMutex.Lock()
	for conn := range s.pHandler.connSet {
		conn.Close()
	}
	s.pHandler.connSet = nil
	s.pHandler.connMutex.Unlock()

	s.pHandler.connWait.Wait()
}
