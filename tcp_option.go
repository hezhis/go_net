package go_net

import "time"

////////////////////////////////////////////
// server

func OptsReadHeadLength(length int) TcpServerOption {
	return func(s *TcpServer) {
		s.param.nHeadLength = length
	}
}

func OptsMaxClientCount(count int) TcpServerOption {
	return func(s *TcpServer) {
		s.nMaxClientCount = count
	}
}

func OptsMaxMsgLength(length uint32) TcpServerOption {
	return func(s *TcpServer) {
		s.param.nMaxMsgLength = length
	}
}

func OptsWriteBuffCap(cap int) TcpServerOption {
	return func(s *TcpServer) {
		s.param.nWriteBuffCap = cap
	}
}

func OptsLittleEndian(flag bool) TcpServerOption {
	return func(s *TcpServer) {
		s.param.bLittleEndian = flag
	}
}

////////////////////////////////////////////
// client

func OptCConnectInterval(interval time.Duration) TcpClientOption {
	return func(c *TcpClient) {
		c.connectInterval = interval
	}
}

func OptCAutoReconnect(b bool) TcpClientOption {
	return func(c *TcpClient) {
		c.autoReconnect = b
	}
}

func OptCReadHeadLength(length int) TcpClientOption {
	return func(c *TcpClient) {
		c.param.nHeadLength = length
	}
}

func OptCMaxMsgLength(length uint32) TcpClientOption {
	return func(s *TcpClient) {
		s.param.nMaxMsgLength = length
	}
}

func OptCWriteBuffCap(cap int) TcpClientOption {
	return func(s *TcpClient) {
		s.param.nWriteBuffCap = cap
	}
}

func OptCLittleEndian(flag bool) TcpClientOption {
	return func(s *TcpClient) {
		s.param.bLittleEndian = flag
	}
}
