package main

import (
	"github.com/hezhis/go_log"
	"github.com/hezhis/go_net"
)

type Agent struct {
	conn go_net.Connector
}

func (agent *Agent) LogicRun() {
	for {
		data, err := agent.conn.ReadMsg()
		if nil != err {
			logger.Error("read message:%v", err)
			break
		}
		logger.Info("read message from[%v] data[%s]", agent.conn.RemoteAddr(), string(data))
		MsgChan <- ChanMsgArgs{
			Id:    "Broadcast",
			Param: data,
		}
	}
}

func (agent *Agent) OnClose() {
	logger.Info("agent OnClose:%v", agent.conn.RemoteAddr())
	MsgChan <- ChanMsgArgs{
		Id:    "CloseAgent",
		Param: agent,
	}
}
