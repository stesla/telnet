package telnet

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoAhead(t *testing.T) {
	in := bytes.NewBuffer([]byte{'h', IAC, GA, 'i'})
	conn := newConnection(in, nil)
	buf := make([]byte, 8)
	n1, err := conn.Read(buf)
	assert.NoError(t, err)
	n2, err := conn.Read(buf[n1:])
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), buf[:n1+n2])
}

func TestNaiveOptions(t *testing.T) {
	var tests = []struct {
		in                   []byte
		expectedr, expectedw []byte
	}{
		{[]byte{'h', IAC, DO, Echo, 'i'}, []byte("hi"), []byte{IAC, WONT, Echo}},
		{[]byte{'h', IAC, DONT, Echo, 'i'}, []byte("hi"), nil},
		{[]byte{'h', IAC, WILL, Echo, 'i'}, []byte("hi"), []byte{IAC, DONT, Echo}},
		{[]byte{'h', IAC, WONT, Echo, 'i'}, []byte("hi"), nil},
	}
	for i, test := range tests {
		in := bytes.NewBuffer(test.in)
		var out bytes.Buffer
		conn := newConnection(in, &out)
		buf := make([]byte, 8)
		var err error
		var n int
		for err == nil {
			var nn int
			nn, err = conn.Read(buf[n:])
			n += nn
		}
		assert.Equal(t, io.EOF, err, "test %d", i)
		assert.Equal(t, test.expectedr, buf[:n], "test %d", i)
		assert.Equal(t, test.expectedw, out.Bytes(), "test %d", i)
	}
}
