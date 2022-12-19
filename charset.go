package telnet

import (
	"bytes"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
)

type CharsetOption struct {
	Option
	enc encoding.Encoding
}

func NewCharsetOption() *CharsetOption {
	return &CharsetOption{Option: NewOption(Charset)}
}

func (c *CharsetOption) Subnegotiation(conn Conn, buf []byte) {
	if len(buf) == 0 {
		conn.Logf(DEBUG, "RECV: IAC SB %s IAC SE", optionByte(c.Byte()))
		return
	}

	cmd, buf := buf[0], buf[1:]

	c.log(conn, "RECV: IAC SB %s %s %s IAC SE", charsetByte(cmd), string(buf))

	if !c.EnabledForUs() {
		c.sendCharsetRejected(conn)
		return
	}

	switch cmd {
	case charsetRequest:
		const ttable = "[TTABLE]"
		if len(buf) > 10 && bytes.HasPrefix(buf, []byte(ttable)) {
			// We don't support TTABLE, so we're just going to strip off the
			// version byte, but according to RFC 2066 it should bsaically always
			// be 0x01. If we ever add TTABLE support, we'll want to check the
			// version to see if it's a version we support.
			buf = buf[len(ttable)+1:]
		}
		if len(buf) < 2 {
			c.sendCharsetRejected(conn)
			return
		}

		charset, encoding := c.selectEncoding(bytes.Split(buf[1:], buf[0:1]))
		if encoding == nil {
			c.sendCharsetRejected(conn)
			return
		} else {
			c.enc = encoding
		}
		c.log(conn, "SEND: IAC SB %s %s %s IAC SE", charsetAccepted, string(charset))
		out := []byte{IAC, SB, Charset, charsetAccepted}
		out = append(out, charset...)
		out = append(out, IAC, SE)
		conn.Send(out)

		them, us := conn.OptionEnabled(TransmitBinary)
		c.Update(conn, TransmitBinary, false, them, false, us)
	}
}

func (c *CharsetOption) Update(conn Conn, option byte, theyChanged, them, weChanged, us bool) {
	switch option {
	case TransmitBinary:
		if c.EnabledForUs() && c.enc != nil {
			if them && us {
				conn.SetEncoding(c.enc)
			} else {
				conn.SetEncoding(ASCII)
			}
		}
	}
}

var encodings = map[string]encoding.Encoding{
	"US-ASCII": ASCII,
}

func (c *CharsetOption) log(conn Conn, fmt string, cmd charsetByte, v ...any) {
	args := []any{
		optionByte(c.Byte()),
		cmd,
	}
	args = append(args, v...)
	conn.Logf(DEBUG, fmt, args...)
}

func (c *CharsetOption) selectEncoding(names [][]byte) (charset []byte, enc encoding.Encoding) {
	for _, name := range names {
		if e, found := encodings[string(name)]; found {
			return name, e
		}

		e, _ := ianaindex.IANA.Encoding(string(name))
		if e != nil {
			return name, e
		}
	}
	return
}

func (c *CharsetOption) sendCharsetRejected(conn Conn) {
	c.log(conn, "SEND: IAC SB %s %s IAC SE", charsetByte(charsetRejected))
	conn.Send([]byte{IAC, SB, Charset, charsetRejected, IAC, SE})
}
