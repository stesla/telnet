package telnet

import (
	"bytes"
	"io"
)

func NewWriter(w io.Writer) io.Writer {
	return &writer{out: w}
}

type writer struct {
	out io.Writer
}

func (w *writer) Write(p []byte) (n int, err error) {
	var buf bytes.Buffer
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
		if _, err := buf.Write(b); err != nil {
			return 0, err
		}
		n++
	}
	_, err = w.out.Write(buf.Bytes())
	return
}
