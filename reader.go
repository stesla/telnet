package telnet

import "io"

const readerCapacity = 1024

func NewReader(r io.Reader) io.Reader {
	result := &reader{
		in:  r,
		buf: make([]byte, readerCapacity),
	}
	result.state = result.decodeByte
	return result
}

type reader struct {
	in     io.Reader
	b, buf []byte
	state  readerState
}

type readerState func(byte) (readerState, byte, bool)

func (r *reader) Read(p []byte) (n int, err error) {
	if len(r.b) == 0 {
		var n int
		n, err = r.in.Read(r.buf)
		r.b = r.buf[:n]
	}
	for len(r.b) > 0 && n < len(p) {
		var c byte
		var ok bool
		r.state, c, ok = r.state(r.b[0])
		r.b = r.b[1:]
		if ok {
			p[n] = c
			n++
		}
	}
	return
}

func (r *reader) decodeByte(c byte) (readerState, byte, bool) {
	switch c {
	case IAC:
		return r.decodeCommand, c, false
	case '\r':
		return r.decodeCarriageReturn, c, false
	default:
		return r.decodeByte, c, true
	}
}

func (r *reader) decodeCommand(c byte) (readerState, byte, bool) {
	switch c {
	case IAC:
		return r.decodeByte, c, true
	default:
		return r.decodeByte, c, false
	}
}

func (r *reader) decodeCarriageReturn(c byte) (readerState, byte, bool) {
	switch c {
	case '\x00':
		return r.decodeByte, '\r', true
	case '\r':
		return r.decodeByte, c, false
	default:
		return r.decodeByte, c, true
	}
}
