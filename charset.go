package telnet

import (
	"bytes"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
)

type CharsetOption struct {
	enabledForThem, enabledForUs bool
	enc                          encoding.Encoding
}

func (*CharsetOption) Option() byte { return Charset }

func (c *CharsetOption) Subnegotiation(conn Conn, buf []byte) {
	cmd, buf := buf[0], buf[1:]

	if !c.enabledForUs {
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
		c.log(conn, charsetAccepted, "SEND: IAC SB %s %s %s IAC SE", string(charset))
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
	case Charset:
		c.enabledForThem, c.enabledForUs = them, us
	case TransmitBinary:
		if c.enabledForUs && c.enc != nil {
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

func (c *CharsetOption) log(conn Conn, cmd charsetByte, fmt string, v ...any) {
	args := []any{
		optionByte(c.Option()),
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
	c.log(conn, charsetByte(charsetRejected), "SEND: IAC SB %s %s IAC SE")
	conn.Send([]byte{IAC, SB, Charset, charsetRejected, IAC, SE})
}
