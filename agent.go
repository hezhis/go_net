package go_net

type Agent interface {
	LogicRun()
	OnClose()
}
