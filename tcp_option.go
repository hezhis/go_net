package go_net

import "time"

////////////////////////////////////////////
// server

func TcpSLocalAddr(addr string) TcpServerOption {
	return func(s *TcpServer) {
		s.sLocalHost = addr
	}
}

func TcpSHeadLen(length int) TcpServerOption {
	return func(s *TcpServer) {
		s.pCreateParam.nHeadLength = length
	}
}

func TcpSMaxClientCount(count int) TcpServerOption {
	return func(s *TcpServer) {
		s.nMaxClientCount = count
	}
}

func TcpSMaxMsgLen(length uint32) TcpServerOption {
	return func(s *TcpServer) {
		s.pCreateParam.nMaxMsgLength = length
	}
}

func TcpSWriteBuffCap(cap int) TcpServerOption {
	return func(s *TcpServer) {
		s.pCreateParam.nWriteBuffCap = cap
	}
}

func TcpSLittleEndian(flag bool) TcpServerOption {
	return func(s *TcpServer) {
		s.pCreateParam.bLittleEndian = flag
	}
}

////////////////////////////////////////////
// client

func TcpCRemoteAddr(addr string) TcpClientOption {
	return func(c *TcpClient) {
		c.sRemoteAddr = addr
	}
}

func TcpCConnectInterval(interval time.Duration) TcpClientOption {
	return func(c *TcpClient) {
		c.nConnectInterval = interval
	}
}

func TcpCAutoReconnect(b bool) TcpClientOption {
	return func(c *TcpClient) {
		c.bAutoReconnect = b
	}
}

func TcpCReadHeadLen(length int) TcpClientOption {
	return func(c *TcpClient) {
		c.pCreateParam.nHeadLength = length
	}
}

func TcpCMaxMsgLen(length uint32) TcpClientOption {
	return func(c *TcpClient) {
		c.pCreateParam.nMaxMsgLength = length
	}
}

func TcpCWriteBuffCap(cap int) TcpClientOption {
	return func(c *TcpClient) {
		c.pCreateParam.nWriteBuffCap = cap
	}
}

func TcpCLittleEndian(flag bool) TcpClientOption {
	return func(c *TcpClient) {
		c.pCreateParam.bLittleEndian = flag
	}
}
