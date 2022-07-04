package telnet

import (
	"io"
	"net"
)

type Conn interface {
	io.Reader
	io.Writer

	AllowOptionForThem(option byte, allow bool)
	AllowOptionForUs(option byte, allow bool)
	EnableOptionForThem(option byte, enable bool) error
	EnableOptionForUs(option byte, enable bool) error
}

func Client(conn net.Conn) Conn {
	return newConnection(conn, conn)
}

func Dial(addr string) (Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return Client(conn), nil
}

func Server(conn net.Conn) Conn {
	return newConnection(conn, conn)
}

type connection struct {
	in     io.Reader
	out    io.Writer
	opts   *optionMap
	rawOut io.Writer
}

func newConnection(r io.Reader, w io.Writer) *connection {
	return &connection{
		in:     NewReader(r),
		opts:   newOptionMap(),
		out:    NewWriter(w),
		rawOut: w,
	}
}

func (c *connection) AllowOptionForThem(option byte, allow bool) {
	opt := c.opts.get(option)
	opt.allowThem = allow
}

func (c *connection) AllowOptionForUs(option byte, allow bool) {
	opt := c.opts.get(option)
	opt.allowUs = allow
}

func (c *connection) EnableOptionForThem(option byte, enable bool) error {
	opt := c.opts.get(option)
	var fn func(sendfunc) error
	if enable {
		fn = opt.enableThem
	} else {
		fn = opt.disableThem
	}
	return fn(func(p ...byte) (err error) {
		_, err = c.rawOut.Write(p)
		return
	})
}

func (c *connection) EnableOptionForUs(option byte, enable bool) error {
	opt := c.opts.get(option)
	var fn func(sendfunc) error
	if enable {
		fn = opt.enableUs
	} else {
		fn = opt.disableUs
	}
	return fn(func(p ...byte) (err error) {
		_, err = c.rawOut.Write(p)
		return
	})
}

func (c *connection) Read(p []byte) (n int, err error) {
	n, err = c.in.Read(p)
	switch t := err.(type) {
	case *telnetGoAhead:
		err = nil
	case *telnetOptionCommand:
		err = nil
		opt := c.opts.get(byte(t.opt))
		opt.receive(byte(t.cmd), func(cmd byte) {
			_, err = c.rawOut.Write([]byte{IAC, cmd, byte(t.opt)})
		})
	}
	return
}

func (c *connection) Write(p []byte) (n int, err error) {
	n, err = c.out.Write(p)
	if err == nil && !c.suppressGoAhead() {
		_, err = c.rawOut.Write([]byte{IAC, GA})
	}
	return
}

func (c *connection) suppressGoAhead() bool {
	return c.opts.get(SuppressGoAhead).enabledForUs()
}
