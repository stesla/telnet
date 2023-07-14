package telnet

import (
	"bytes"
	"io"
	"net"
	"time"
)

//go:generate mockgen -package=telnet -destination=test_mocks.go . Conn,Option,Logger,EventSink

type testConn struct {
	io.Reader
	io.Writer
}

func newTestConn(r io.Reader, w io.Writer) *connection {
	if r == nil {
		var buf bytes.Buffer
		r = &buf
	}
	if w == nil {
		var buf bytes.Buffer
		w = &buf
	}
	return newConnection(&testConn{r, w})
}

func (t *testConn) Close() error                     { return nil }
func (t *testConn) LocalAddr() net.Addr              { return nil }
func (t *testConn) RemoteAddr() net.Addr             { return nil }
func (t *testConn) SetDeadline(time.Time) error      { return nil }
func (t *testConn) SetReadDeadline(time.Time) error  { return nil }
func (t *testConn) SetWriteDeadline(time.Time) error { return nil }
