package telnet

type EventListener interface {
	HandleEvent(any)
}

type FuncListener struct {
	Func func(any)
}

func (f FuncListener) HandleEvent(data any) { f.Func(data) }

type EventSink interface {
	SendEvent(event string, data any)
}
