package telnet

import (
	"io"
	"net"

	"golang.org/x/text/encoding"
)

type Conn interface {
	io.Reader
	io.Writer

	AllowOption(handler OptionHandler, allowThem, allowUs bool)
	EnableOptionForThem(option byte, enable bool) error
	EnableOptionForUs(option byte, enable bool) error
	OptionEnabled(option byte) (them, us bool)

	Send(p []byte) (n int, err error)
	SetEncoding(encoding.Encoding)
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

type OptionHandler interface {
	Option() byte
	Enable(Conn)
	Disable(Conn)
	Subnegotiation(Conn, []byte)
}

type connection struct {
	opts            *optionMap
	handlers        map[byte]OptionHandler
	r, in           io.Reader
	w, out          io.Writer
	suppressGoAhead bool
}

func newConnection(r io.Reader, w io.Writer) *connection {
	conn := &connection{
		opts:     newOptionMap(),
		handlers: make(map[byte]OptionHandler),
		r:        r,
		w:        w,
	}
	conn.SetEncoding(ASCII)
	return conn
}

func (c *connection) AllowOption(handler OptionHandler, allowThem, allowUs bool) {
	opt := c.opts.get(handler.Option())
	c.handlers[opt.code], opt.allowThem, opt.allowUs = handler, allowThem, allowUs
}

func (c *connection) EnableOptionForThem(option byte, enable bool) error {
	opt := c.opts.get(option)
	var fn func(sendfunc) error
	if enable {
		fn = opt.enableThem
	} else {
		fn = opt.disableThem
	}
	return fn(func(p []byte) (err error) {
		_, err = c.Send(p)
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
	return fn(func(p []byte) (err error) {
		_, err = c.Send(p)
		return
	})
}

func (c *connection) OptionEnabled(option byte) (them, us bool) {
	opt := c.opts.get(option)
	them, us = opt.enabledForThem(), opt.enabledForUs()
	return
}

func (c *connection) Read(p []byte) (n int, err error) {
	return c.in.Read(p)
}

func (c *connection) Send(p []byte) (int, error) {
	return c.w.Write(p)
}

func (c *connection) SetEncoding(enc encoding.Encoding) {
	c.in = enc.NewDecoder().Reader(NewReader(c.r, c.handleCommand))
	c.out = enc.NewEncoder().Writer(NewWriter(c.w))
}

func (c *connection) SuppressGoAhead(enabled bool) {
	c.suppressGoAhead = enabled
}

func (c *connection) Write(p []byte) (n int, err error) {
	n, err = c.out.Write(p)
	if err == nil && !c.suppressGoAhead {
		_, err = c.Send([]byte{IAC, GA})
	}
	return
}

func (c *connection) handleCommand(cmd any) (err error) {
	switch t := cmd.(type) {
	case *telnetGoAhead:
		// do nothing
	case *telnetOptionCommand:
		opt := c.opts.get(byte(t.opt))
		enabled := opt.enabledForUs()
		err = opt.receive(t.cmd, func(p []byte) (err error) {
			_, err = c.Send(p)
			return
		})
		if handler, ok := c.handlers[opt.code]; ok {
			if !enabled && opt.enabledForUs() {
				handler.Enable(c)
			} else if enabled && !opt.enabledForUs() {
				handler.Disable(c)
			}
		}
	case *telnetSubnegotiation:
		if handler, ok := c.handlers[t.opt]; ok && len(t.bytes) > 0 {
			handler.Subnegotiation(c, t.bytes)
		}
	}
	return
}

type SuppressGoAheadOption struct{}

func (SuppressGoAheadOption) Option() byte { return SuppressGoAhead }

func (SuppressGoAheadOption) Enable(conn Conn) {
	conn.SuppressGoAhead(true)
}

func (SuppressGoAheadOption) Disable(conn Conn) {
	conn.SuppressGoAhead(false)
}

func (SuppressGoAheadOption) Subnegotiation(Conn, []byte) {}
