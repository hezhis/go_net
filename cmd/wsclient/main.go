package main

import (
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/hezhis/go_log"
	"github.com/hezhis/go_net"
)

func main() {
	rand.Seed(time.Now().UnixMicro())
	client := go_net.NewWSClient(
		go_net.WSCRemoteAddr("ws://127.0.0.1:6321"),
		go_net.WSCAutoReconnect(true),
	)

	var agent *Agent
	client.NewAgent = func(connector *go_net.WSConnector) go_net.Agent {
		agent = &Agent{conn: connector, name: strconv.Itoa(rand.Int())}
		logger.Info("new agent remote addr:%v", agent.conn.LocalAddr())
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
			logger.Info("close by signal:%v", sig)
			break out
		case <-ticker.C:
			if nil != agent && nil != agent.conn {
				agent.conn.WriteMsg([]byte("my name's " + agent.name))
			}
		}
	}

	client.Close()
}
