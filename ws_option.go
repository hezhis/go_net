package go_net

import "time"

func WSSLocalAddr(addr string) WSServerOption {
	return func(s *WSServer) {
		s.sLocalHost = addr
	}
}

func WSSMaxClientCount(count int) WSServerOption {
	return func(s *WSServer) {
		s.nMaxClientCount = count
	}
}

func WSSHTTPTimeout(t time.Duration) WSServerOption {
	return func(s *WSServer) {
		s.nHTTPTimeout = t
	}
}

func WSSCertFile(f string) WSServerOption {
	return func(s *WSServer) {
		s.sCertFile = f
	}
}

func WSSKeyFile(f string) WSServerOption {
	return func(s *WSServer) {
		s.sKeyFile = f
	}
}

func WSCAutoReconnect(flag bool) WSClientOption {
	return func(c *WSClient) {
		c.bAutoReconnect = flag
	}
}

////////////////////////////////////////////
// client

func WSCRemoteAddr(addr string) WSClientOption {
	return func(c *WSClient) {
		c.sRemoteAddr = addr
	}
}

func WSCConnectInterval(interval time.Duration) WSClientOption {
	return func(c *WSClient) {
		c.nConnectInterval = interval
	}
}
