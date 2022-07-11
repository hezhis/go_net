package main

type (
	ChanMsgArgs struct {
		Id    string
		Param interface{}
	}
)

var (
	AgentSet = make(map[*Agent]struct{})
	MsgChan  = make(chan ChanMsgArgs, 100)
)
