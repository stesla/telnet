package telnet

import "io"

func NewWriter(w io.Writer) io.Writer {
	return &writer{out: w}
}

type writer struct {
	out io.Writer
}

func (w *writer) Write(p []byte) (n int, err error) {
	for _, c := range p {
		var b []byte
		switch c {
		case '\n':
			b = []byte("\r\n")
		case '\r':
			b = []byte("\r\x00")
		case IAC:
			b = []byte{IAC, IAC}
		default:
			b = []byte{c}
		}
		_, err = w.out.Write(b)
		if err != nil {
			return
		}
		n++
	}
	return
}
