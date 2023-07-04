package telnet

import (
	"errors"
	"fmt"
	"io"
	"net"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
)

type Conn interface {
	io.Reader
	io.Writer
	Logger

	BindOption(o Option)
	EnableOptionForThem(option byte, enable bool) error
	EnableOptionForUs(option byte, enable bool) error
	OptionEnabled(option byte) (them, us bool)

	RequestEncoding(encoding.Encoding) error
	Send(p []byte) (n int, err error)
	SetEncoding(encoding.Encoding)
	SetLogger(Logger)
	SetReadEncoding(encoding.Encoding)
	SetWriteEncoding(encoding.Encoding)
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
	Logger

	opts            *optionMap
	r, in           io.Reader
	w, out          io.Writer
	suppressGoAhead bool
}

func newConnection(r io.Reader, w io.Writer) *connection {
	conn := &connection{
		Logger: NullLogger{},
		opts:   newOptionMap(),
		r:      r,
		w:      w,
	}
	conn.opts.each(func(o Option) { o.Bind(conn) })
	conn.SetEncoding(ASCII)
	return conn
}

func (c *connection) BindOption(o Option) {
	o.Bind(c)
	c.opts.put(o)
}

func (c *connection) EnableOptionForThem(option byte, enable bool) error {
	opt := c.opts.get(option)
	var fn func() error
	if enable {
		fn = opt.enableThem
	} else {
		fn = opt.disableThem
	}
	return fn()
}

func (c *connection) EnableOptionForUs(option byte, enable bool) error {
	opt := c.opts.get(option)
	var fn func() error
	if enable {
		fn = opt.enableUs
	} else {
		fn = opt.disableUs
	}
	return fn()
}

func (c *connection) OptionEnabled(option byte) (them, us bool) {
	opt := c.opts.get(option)
	them, us = opt.EnabledForThem(), opt.EnabledForUs()
	return
}

func (c *connection) Read(p []byte) (n int, err error) {
	return c.in.Read(p)
}

func (c *connection) RequestEncoding(enc encoding.Encoding) error {
	if _, us := c.OptionEnabled(Charset); !us {
		return errors.New("charset option not enabled")
	}
	msg := []byte{IAC, SB, charsetRequest}
	str, err := ianaindex.IANA.Name(enc)
	if err != nil {
		return err
	}
	msg = append(msg, str...)
	msg = append(msg, SE)

	_, err = c.Send(msg)
	return err
}

func (c *connection) Send(p []byte) (int, error) {
	return c.w.Write(p)
}

func (c *connection) SetEncoding(enc encoding.Encoding) {
	c.SetReadEncoding(enc)
	c.SetWriteEncoding(enc)
}

func (c *connection) SetLogger(logger Logger) {
	c.Logger = logger
}

func (c *connection) SetReadEncoding(enc encoding.Encoding) {
	c.in = enc.NewDecoder().Reader(NewReader(c.r, c.handleCommand))
}

func (c *connection) SetWriteEncoding(enc encoding.Encoding) {

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
	if s, ok := cmd.(fmt.Stringer); ok {
		c.Logf(DEBUG, "RECV: %s", s)
	}

	switch t := cmd.(type) {
	case *telnetGoAhead:
		// do nothing
	case *telnetOptionCommand:
		opt := c.opts.get(byte(t.opt))
		them, us := opt.EnabledForThem(), opt.EnabledForUs()
		err = opt.receive(t.cmd)
		if err != nil {
			return
		}
		newThem, newUs := opt.EnabledForThem(), opt.EnabledForUs()
		theyChanged := them != newThem
		weChanged := us != newUs
		c.opts.each(func(o Option) {
			o.Update(opt.Byte(), theyChanged, newThem, weChanged, newUs)
		})
	case *telnetSubnegotiation:
		option := c.opts.get(t.opt)
		option.Subnegotiation(t.bytes)
	}
	return
}

func (c *connection) sendOptionCommand(cmd, opt byte) (err error) {
	c.Logf(DEBUG, "SEND: IAC %s %s", commandByte(cmd), optionByte(opt))
	_, err = c.Send([]byte{IAC, cmd, opt})
	return
}

type SuppressGoAheadOption struct {
	Option
}

func NewSuppressGoAheadOption() *SuppressGoAheadOption {
	return &SuppressGoAheadOption{Option: NewOption(SuppressGoAhead)}
}

func (o *SuppressGoAheadOption) Subnegotiation([]byte) {}

func (o *SuppressGoAheadOption) Update(option byte, theyChanged, them, weChanged, us bool) {
	if SuppressGoAhead == option && weChanged {
		o.Conn().SuppressGoAhead(us)
	}
}
