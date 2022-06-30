package telnet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteAscii(t *testing.T) {
	var tests = []struct {
		in, expected []byte
	}{
		{[]byte("hello"), []byte("hello")},
		{[]byte{'h', IAC, 'i'}, []byte{'h', IAC, IAC, 'i'}},
		{[]byte("foo\nbar"), []byte("foo\r\nbar")},
		{[]byte("foo\rbar"), []byte("foo\r\x00bar")},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		w := NewWriter(&buf)
		n, err := w.Write(test.in)
		assert.NoError(t, err)
		assert.Equal(t, len(test.in), n)
		assert.Equal(t, test.expected, buf.Bytes())
	}
}
