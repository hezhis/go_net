package main

import (
	"github.com/hezhis/go_net"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	client := go_net.NewTcpClient(
		go_net.TcpCRemoteAddr("127.0.0.1:6321"),
		go_net.TcpCAutoReconnect(true),
	)
	var agent *Agent
	client.NewAgent = func(connector *go_net.TcpConnector) go_net.Agent {
		agent = &Agent{conn: connector}
		log.Printf("new agent remote addr:%v", agent.conn.LocalAddr())
		return agent
	}
	client.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
out:
	for {
		select {
		case sig := <-c:
			log.Printf("close by signal:%v", sig)
			break out
		case <-ticker.C:
			if nil != agent && nil != agent.conn {
				agent.conn.WriteMsg([]byte{0, 1, 2, 3, 4})
			}
		}
	}

	client.Close()
}
