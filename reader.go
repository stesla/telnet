package telnet

import (
	"fmt"
	"io"
)

func NewReader(r io.Reader) io.Reader {
	result := &reader{in: r}
	result.state = result.decodeByte
	return result
}

type reader struct {
	in     io.Reader
	b, buf []byte
	state  readerState
}

type readerState func(byte) readerStateTransition
type readerStateTransition struct {
	state readerState
	c     byte
	ok    bool
	err   error
}

func (r *reader) Read(p []byte) (n int, err error) {
	if len(r.b) == 0 {
		var n int
		r.b = make([]byte, len(p))
		n, err = r.in.Read(r.b)
		r.b = r.b[:n]
	}
	for len(r.b) > 0 && n < len(p) {
		t := r.state(r.b[0])
		r.b = r.b[1:]
		r.state = t.state
		if t.ok {
			p[n] = t.c
			n++
		}
		if t.err != nil {
			return n, t.err
		}
	}
	return
}

func (r *reader) decodeByte(c byte) readerStateTransition {
	switch c {
	case IAC:
		return readerStateTransition{state: r.decodeCommand, c: c, ok: false}
	case '\r':
		return readerStateTransition{state: r.decodeCarriageReturn, c: c, ok: false}
	default:
		return readerStateTransition{state: r.decodeByte, c: c, ok: true}
	}
}

func (r *reader) decodeCommand(c byte) readerStateTransition {
	switch c {
	case IAC:
		return readerStateTransition{state: r.decodeByte, c: c, ok: true}
	case DO, DONT, WILL, WONT:
		return readerStateTransition{state: r.decodeOption(c), c: c, ok: false}
	case GA:
		return readerStateTransition{state: r.decodeByte, c: c, ok: false, err: &telnetGoAhead{}}
	case SB:
		return readerStateTransition{state: r.decodeSubnegotiation(), c: c, ok: false}
	default:
		return readerStateTransition{state: r.decodeByte, c: c, ok: false}
	}
}

func (r *reader) decodeCarriageReturn(c byte) readerStateTransition {
	switch c {
	case '\x00':
		return readerStateTransition{state: r.decodeByte, c: '\r', ok: true}
	case '\r':
		return readerStateTransition{state: r.decodeByte, c: c, ok: false}
	default:
		return readerStateTransition{state: r.decodeByte, c: c, ok: true}
	}
}

func (r *reader) decodeOption(cmd byte) readerState {
	return func(c byte) readerStateTransition {
		err := &telnetOptionCommand{commandByte(cmd), optionByte(c)}
		return readerStateTransition{state: r.decodeByte, c: c, ok: false, err: err}
	}
}

const subnegotiationBufferSize = 256

func (r *reader) decodeSubnegotiation() readerState {
	var buf = make([]byte, 0, subnegotiationBufferSize)

	var readByte, seenIAC readerState

	readByte = func(c byte) readerStateTransition {
		switch c {
		case IAC:
			return readerStateTransition{state: seenIAC, c: c, ok: false}
		default:
			buf = append(buf, c)
			return readerStateTransition{state: readByte, c: c, ok: false}
		}
	}

	seenIAC = func(c byte) readerStateTransition {
		switch c {
		case IAC:
			buf = append(buf, c)
			return readerStateTransition{state: readByte, c: c, ok: false}
		case SE:
			var err error
			if len(buf) > 0 {
				err = &telnetSubnegotiation{buf}
			}
			return readerStateTransition{state: r.decodeByte, c: c, ok: false, err: err}
		default:
			err := fmt.Errorf("IAC %b", commandByte(c))
			return readerStateTransition{state: r.decodeByte, c: c, ok: false, err: err}
		}
	}

	return readByte
}
