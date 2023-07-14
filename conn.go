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
	net.Conn
	Logger

	AddListener(string, EventListener)
	RemoveListener(string, EventListener)

	BindOption(o Option)
	EnableOptionForThem(option byte, enable bool) error
	EnableOptionForUs(option byte, enable bool) error
	Option(option byte) Option

	RequestEncoding(encoding.Encoding) error
	Send(p []byte) (n int, err error)
	SetEncoding(encoding.Encoding)
	SetLogger(Logger)
	SetReadEncoding(encoding.Encoding)
	SetWriteEncoding(encoding.Encoding)
	SuppressGoAhead(enabled bool)
}

func Client(conn net.Conn) Conn {
	return newConnection(conn)
}

func Dial(addr string) (Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return Client(conn), nil
}

func Server(conn net.Conn) Conn {
	return newConnection(conn)
}

type connection struct {
	net.Conn
	Logger

	listeners       map[string][]EventListener
	opts            *optionMap
	in              io.Reader
	out             io.Writer
	suppressGoAhead bool
}

func newConnection(upstream net.Conn) *connection {
	conn := &connection{
		Conn:      upstream,
		Logger:    NullLogger{},
		listeners: map[string][]EventListener{},
		opts:      newOptionMap(),
	}
	conn.opts.each(func(o Option) { o.Bind(conn, conn) })
	conn.SetEncoding(ASCII)
	return conn
}

func (c *connection) AddListener(event string, l EventListener) {
	c.listeners[event] = append(c.listeners[event], l)
}

func (c *connection) BindOption(o Option) {
	o.Bind(c, c)
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

func (c *connection) Option(option byte) Option {
	return c.opts.get(option)
}

func (c *connection) Read(p []byte) (n int, err error) {
	return c.in.Read(p)
}

func (c *connection) RemoveListener(event string, l EventListener) {
	var i int
	listeners := c.listeners[event]
	for i = range listeners {
		if l == listeners[i] {
			c.listeners[event] = append(listeners[:i], listeners[i+1:]...)
			return
		}
	}
}

func (c *connection) RequestEncoding(enc encoding.Encoding) error {
	if opt := c.Option(Charset); !opt.EnabledForUs() {
		return errors.New("charset option not enabled")
	}
	msg := []byte{IAC, SB, Charset, charsetRequest, ';'}
	str, err := ianaindex.IANA.Name(enc)
	if err != nil {
		return err
	}
	msg = append(msg, str...)
	msg = append(msg, IAC, SE)

	c.Logf("SEND: IAC SB %s %s ;%s IAC SE", optionByte(Charset), charsetByte(charsetRequest), str)
	_, err = c.Send(msg)
	return err
}

func (c *connection) Send(p []byte) (int, error) {
	return c.Conn.Write(p)
}

func (c *connection) SendEvent(event string, data any) {
	for _, l := range c.listeners[event] {
		l.HandleEvent(data)
	}
}

func (c *connection) SetEncoding(enc encoding.Encoding) {
	c.SetReadEncoding(enc)
	c.SetWriteEncoding(enc)
}

func (c *connection) SetLogger(logger Logger) {
	c.Logger = logger
}

func (c *connection) SetReadEncoding(enc encoding.Encoding) {
	c.in = enc.NewDecoder().Reader(NewReader(c.Conn, c.handleCommand))
}

func (c *connection) SetWriteEncoding(enc encoding.Encoding) {

	c.out = enc.NewEncoder().Writer(NewWriter(c.Conn))
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
		c.Logf("RECV: %s", s)
	}

	switch t := cmd.(type) {
	case *telnetGoAhead:
		// do nothing
	case *telnetOptionCommand:
		opt := c.opts.get(byte(t.opt))
		err = opt.receive(t.cmd)
		if err != nil {
			return
		}
	case *telnetSubnegotiation:
		option := c.opts.get(t.opt)
		option.Subnegotiation(t.bytes)
	}
	return
}

type SuppressGoAheadOption struct {
	Option
}

func NewSuppressGoAheadOption() *SuppressGoAheadOption {
	return &SuppressGoAheadOption{Option: NewOption(SuppressGoAhead)}
}

func (o *SuppressGoAheadOption) Bind(conn Conn, sink EventSink) {
	o.Option.Bind(conn, sink)
	conn.AddListener("update-option", o)
}

func (o *SuppressGoAheadOption) Subnegotiation([]byte) {}

func (o *SuppressGoAheadOption) HandleEvent(data any) {
	event, ok := data.(UpdateOptionEvent)
	if !ok {
		return
	}

	if SuppressGoAhead == event.Option.Byte() && event.WeChanged {
		o.Conn().SuppressGoAhead(event.Option.EnabledForUs())
	}
}

type Logger interface {
	Logf(fmt string, v ...any)
}

type NullLogger struct{}

func (NullLogger) Logf(string, ...any) {}
