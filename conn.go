package telnet

import (
	"io"
	"net"
)

type Conn interface {
	io.Reader
	io.Writer

	AllowOption(handler OptionHandler, allowThem, allowUs bool)
	EnableOptionForThem(option byte, enable bool) error
	EnableOptionForUs(option byte, enable bool) error

	Send(p ...byte) (n int, err error)
	SuppressGoAhead(enabled bool)
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

	suppressGoAhead bool
}

func newConnection(r io.Reader, w io.Writer) *connection {
	return &connection{
		in:     NewReader(r),
		opts:   newOptionMap(),
		out:    NewWriter(w),
		rawOut: w,
	}
}

func (c *connection) AllowOption(handler OptionHandler, allowThem, allowUs bool) {
	opt := c.opts.get(handler.Option())
	opt.OptionHandler, opt.allowThem, opt.allowUs = handler, allowThem, allowUs
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
		_, err = c.Send(p...)
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
		_, err = c.Send(p...)
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
			if opt.OptionHandler != nil {
				switch cmd {
				case DO:
					opt.Enable(c)
				case DONT:
					opt.Disable(c)
				}
			}
			_, err = c.Send(IAC, cmd, byte(t.opt))
		})
	}
	return
}

func (c *connection) Send(p ...byte) (int, error) {
	return c.rawOut.Write(p)
}

func (c *connection) SuppressGoAhead(enabled bool) {
	c.suppressGoAhead = enabled
}

func (c *connection) Write(p []byte) (n int, err error) {
	n, err = c.out.Write(p)
	if err == nil && !c.suppressGoAhead {
		_, err = c.Send(IAC, GA)
	}
	return
}
