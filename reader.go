package telnet

import (
	"io"
)

func NewReader(r io.Reader, fn func(any) error) io.Reader {
	result := &reader{in: r, cmdfn: fn}
	result.state = result.decodeByte
	return result
}

type reader struct {
	in    io.Reader
	b     []byte
	state readerState
	cmdfn func(any) error
}

type readerState func(byte) (readerState, byte, bool, error)

func (r *reader) Read(p []byte) (n int, err error) {
	if len(r.b) == 0 {
		var n int
		r.b = make([]byte, len(p))
		n, err = r.in.Read(r.b)
		r.b = r.b[:n]
	}
	for len(r.b) > 0 && n < len(p) {
		state, c, ok, err := r.state(r.b[0])
		r.b = r.b[1:]
		r.state = state
		if ok {
			p[n] = c
			n++
		}
		if err != nil {
			return n, err
		}
	}
	return
}

func (r *reader) decodeByte(c byte) (readerState, byte, bool, error) {
	switch c {
	case IAC:
		return r.decodeCommand, c, false, nil
	case '\r':
		return r.decodeCarriageReturn, c, false, nil
	default:
		return r.decodeByte, c, true, nil
	}
}

func (r *reader) decodeCommand(c byte) (readerState, byte, bool, error) {
	switch c {
	case IAC:
		return r.decodeByte, c, true, nil
	case DO, DONT, WILL, WONT:
		return r.decodeOption(c), c, false, nil
	case GA:
		err := r.handleCommand(&telnetGoAhead{})
		return r.decodeByte, c, false, err
	case SB:
		return r.decodeSubnegotiation, c, false, nil
	default:
		return r.decodeByte, c, false, nil
	}
}

func (r *reader) decodeCarriageReturn(c byte) (readerState, byte, bool, error) {
	switch c {
	case '\x00':
		return r.decodeByte, '\r', true, nil
	case '\r':
		return r.decodeByte, c, false, nil
	default:
		return r.decodeByte, c, true, nil
	}
}

func (r *reader) decodeOption(cmd byte) readerState {
	return func(c byte) (readerState, byte, bool, error) {
		err := r.handleCommand(&telnetOptionCommand{cmd, c})
		return r.decodeByte, c, false, err
	}
}

const subnegotiationBufferSize = 256

func (r *reader) decodeSubnegotiation(option byte) (readerState, byte, bool, error) {
	var buf = make([]byte, 0, subnegotiationBufferSize)

	var readByte, seenIAC readerState

	readByte = func(c byte) (readerState, byte, bool, error) {
		switch c {
		case IAC:
			return seenIAC, c, false, nil
		default:
			buf = append(buf, c)
			return readByte, c, false, nil
		}
	}

	seenIAC = func(c byte) (readerState, byte, bool, error) {
		switch c {
		case IAC:
			buf = append(buf, c)
			return readByte, c, false, nil
		case SE:
			var err error
			if len(buf) > 0 {
				err = r.handleCommand(&telnetSubnegotiation{option, buf})
			}
			return r.decodeByte, c, false, err
		default:
			return r.decodeByte, c, false, nil
		}
	}

	return readByte, option, false, nil
}

func (r *reader) handleCommand(cmd any) (err error) {
	if r.cmdfn != nil {
		err = r.cmdfn(cmd)
	}
	return
}
