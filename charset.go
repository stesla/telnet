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

func (c *CharsetOption) Bind(conn Conn, sink EventSink) {
	c.Option.Bind(conn, sink)
	conn.AddListener(c)
}

func (c *CharsetOption) Subnegotiation(buf []byte) {
	if len(buf) == 0 {
		c.log("RECV: IAC SB %s IAC SE", optionByte(c.Byte()))
		return
	}

	cmd, buf := buf[0], buf[1:]
	c.logCharsetCommand("RECV: IAC SB %s %s %s IAC SE", charsetByte(cmd), string(buf))

	switch cmd {
	case charsetAccepted:
		c.enc = c.getEncoding(buf)
		c.updateWithBinaryStatus()

	case charsetRejected:
		c.Sink().SendEvent("charset-rejected", nil)

	case charsetRequest:
		if !c.EnabledForUs() {
			c.sendCharsetRejected()
			return
		}

		const ttable = "[TTABLE]"
		if len(buf) > 10 && bytes.HasPrefix(buf, []byte(ttable)) {
			// We don't support TTABLE, so we're just going to strip off the
			// version byte, but according to RFC 2066 it should basically always
			// be 0x01. If we ever add TTABLE support, we'll want to check the
			// version to see if it's a version we support.
			buf = buf[len(ttable)+1:]
		}
		if len(buf) < 2 {
			c.sendCharsetRejected()
			return
		}

		charset, encoding := c.selectEncoding(bytes.Split(buf[1:], buf[0:1]))
		if encoding == nil {
			c.sendCharsetRejected()
			return
		} else {
			c.enc = encoding
		}
		c.logCharsetCommand("SEND: IAC SB %s %s %s IAC SE", charsetAccepted, string(charset))
		out := []byte{IAC, SB, Charset, charsetAccepted}
		out = append(out, charset...)
		out = append(out, IAC, SE)
		c.send(out)
		c.updateWithBinaryStatus()
	case charsetTTableIs:
		// We don't support TTABLE, but we don't want to leave our peers hanging
		// if they send us a TTABLE-IS subnegotiation.
		c.Conn().Send([]byte{IAC, SB, Charset, charsetTTableRejected, IAC, SE})
	}
}

func (c *CharsetOption) HandleEvent(name string, data any) {
	event, ok := data.(UpdateOptionEvent)
	if !ok {
		return
	}

	switch opt := event.Option; opt.Byte() {
	case TransmitBinary:
		if c.EnabledForUs() && c.enc != nil {
			conn := c.Conn()
			sink := c.Sink()
			if opt.EnabledForThem() && opt.EnabledForUs() {
				conn.SetEncoding(c.enc)
				sink.SendEvent("charset-accepted", c.enc)
			} else {
				conn.SetEncoding(ASCII)
			}
		}
	}
}

var encodings = map[string]encoding.Encoding{
	"US-ASCII": ASCII,
}

func (c *CharsetOption) log(fmt string, args ...any) {
	c.Conn().Logf(fmt, args...)
}

func (c *CharsetOption) logCharsetCommand(fmt string, cmd charsetByte, v ...any) {
	args := []any{
		optionByte(c.Byte()),
		cmd,
	}
	args = append(args, v...)
	c.log(fmt, args...)
}

func (c *CharsetOption) selectEncoding(names [][]byte) (charset []byte, enc encoding.Encoding) {
	for _, name := range names {
		charset := c.getEncoding(name)
		if charset != nil {
			return name, charset
		}
	}
	return
}

func (c *CharsetOption) updateWithBinaryStatus() {
	c.HandleEvent("update-option", UpdateOptionEvent{
		c.Conn().Option(TransmitBinary),
		false,
		false,
	})
}

func (*CharsetOption) getEncoding(name []byte) encoding.Encoding {
	if e, found := encodings[string(name)]; found {
		return e
	}

	e, _ := ianaindex.IANA.Encoding(string(name))
	if e != nil {
		return e
	}

	return nil
}

func (c *CharsetOption) send(p []byte) {
	c.Conn().Send(p)
}

func (c *CharsetOption) sendCharsetRejected() {
	c.logCharsetCommand("SEND: IAC SB %s %s IAC SE", charsetByte(charsetRejected))
	c.send([]byte{IAC, SB, Charset, charsetRejected, IAC, SE})
}
