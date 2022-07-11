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
		log.Printf("read message:%v", data)
	}
}

func (agent *Agent) OnClose() {
	log.Printf("agent OnClose:%v", agent.conn.RemoteAddr())
}
