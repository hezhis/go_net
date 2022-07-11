package main

import (
	"github.com/hezhis/go_net"
	"log"
	"os"
	"os/signal"
)

func main() {
	s := go_net.NewWSServer(
		go_net.WSSLocalAddr("127.0.0.1:6321"),
		go_net.WSSMaxClientCount(2),
	)

	s.NewAgent = func(connector *go_net.WSConnector) go_net.Agent {
		agent := &Agent{conn: connector}
		log.Printf("new agent remote addr:%v", agent.conn.RemoteAddr())
		MsgChan <- ChanMsgArgs{
			Id:    "NewAgent",
			Param: agent,
		}
		return agent
	}

	s.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

out:
	for {
		select {
		case sig := <-c:
			break out
			log.Printf("close by signal:%v", sig)
		case msg := <-MsgChan:
			switch msg.Id {
			case "NewAgent":
				if agent, ok := msg.Param.(*Agent); ok {
					AgentSet[agent] = struct{}{}
				}
			case "CloseAgent":
				if agent, ok := msg.Param.(*Agent); ok {
					delete(AgentSet, agent)
				}
			case "Broadcast":
				if buf, ok := msg.Param.([]byte); ok {
					for agent := range AgentSet {
						agent.conn.WriteMsg(buf)
					}
				}
			}
		}
	}

	s.Close()
}
