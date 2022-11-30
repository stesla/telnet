package telnet

import (
	"bytes"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
)

type CharsetOption struct{}

func (*CharsetOption) Option() byte        { return Charset }
func (*CharsetOption) EnableForUs(Conn)    {}
func (*CharsetOption) EnableForThem(Conn)  {}
func (*CharsetOption) DisableForUs(Conn)   {}
func (*CharsetOption) DisableForThem(Conn) {}

func (c *CharsetOption) Subnegotiation(conn Conn, buf []byte) {
	cmd, buf := buf[0], buf[1:]
	switch cmd {
	case charsetRequest:
		const ttable = "[TTABLE]"
		if len(buf) > 10 && bytes.HasPrefix(buf, []byte(ttable)) {
			// strip off the version byte
			buf = buf[len(ttable)+1:]
		}
		// if len(buf) < 2 {
		// 	p.sendCharsetRejected()
		// 	return
		// }

		charset, encoding := c.selectEncoding(bytes.Split(buf[1:], buf[0:1]))
		// if encoding == nil {
		// 	p.sendCharsetRejected()
		// 	return
		// }
		out := []byte{IAC, SB, Charset, charsetAccepted}
		out = append(out, charset...)
		out = append(out, IAC, SE)
		conn.Send(out)
		conn.SetEncoding(encoding)
	}
}

func (c *CharsetOption) selectEncoding(names [][]byte) (charset []byte, enc encoding.Encoding) {
	for _, name := range names {
		switch string(name) {
		case "UTF-8":
			return name, unicode.UTF8
		case "US-ASCII":
			return name, ASCII
		}
	}
	return
}
