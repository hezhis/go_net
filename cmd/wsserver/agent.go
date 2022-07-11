package main

import (
	"github.com/hezhis/go_net"
	"log"
)

type Agent struct {
	conn go_net.Connector
}

func (agent *Agent) LogicRun() {
	for {
		data, err := agent.conn.ReadMsg()
		if nil != err {
			log.Printf("read message:%v", err)
			break
		}
		log.Printf("read message from[%v] data[%s]", agent.conn.RemoteAddr(), string(data))
		MsgChan <- ChanMsgArgs{
			Id:    "Broadcast",
			Param: data,
		}
	}
}

func (agent *Agent) OnClose() {
	log.Printf("agent OnClose:%v", agent.conn.RemoteAddr())
	MsgChan <- ChanMsgArgs{
		Id:    "CloseAgent",
		Param: agent,
	}
}
