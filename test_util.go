package telnet

//go:generate mockgen -package=telnet -destination=test_mocks.go . Conn,Option,Logger,EventSink
