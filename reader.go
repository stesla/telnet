package telnet

import "io"

const readerCapacity = 1024

func NewReader(r io.Reader) io.Reader {
	return &reader{
		in:  r,
		buf: make([]byte, readerCapacity),
	}
}

type reader struct {
	in      io.Reader
	b, buf  []byte
	cr, iac bool
}

func (r *reader) Read(p []byte) (n int, err error) {
	if len(r.b) == 0 {
		var n int
		n, err = r.in.Read(r.buf)
		r.b = r.buf[:n]
	}
	for len(r.b) > 0 && n < len(p) {
		c := r.b[0]
		r.b = r.b[1:]
		if r.iac {
			r.iac = false
			if c == IAC {
				p[n] = IAC
				n++
			}
			continue
		} else if r.cr {
			r.cr = false
			if c == 0 {
				p[n] = '\r'
				n++
				continue
			}
		}
		switch c {
		case IAC:
			r.iac = true
		case '\r':
			r.cr = true
		default:
			p[n] = c
			n++
		}
	}
	return
}
