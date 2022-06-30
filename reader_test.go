package telnet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readTest(in []byte) (out []byte, err error) {
	r := NewReader(bytes.NewBuffer(in))
	out = make([]byte, 16)
	n, err := r.Read(out)
	out = out[:n]
	return
}

func TestSimple(t *testing.T) {
	var tests = []struct {
		in, expected []byte
	}{
		{[]byte("hello"), []byte("hello")},
		{[]byte{'h', IAC, NOP, 'i'}, []byte("hi")},
		{[]byte{'h', IAC, IAC, 'i'}, []byte{'h', IAC, 'i'}},
		{[]byte("foo\r\nbar"), []byte("foo\nbar")},
		{[]byte("foo\r\x00bar"), []byte("foo\rbar")},
	}
	for _, test := range tests {
		buf, err := readTest(test.in)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, buf)
	}
}

func TestCRIsUsuallyIgnored(t *testing.T) {
	const (
		min byte = 0
		max      = 127
	)
	for c := min; c < max; c++ {
		if c == 0 || c == '\n' {
			continue
		}
		buf, err := readTest([]byte{'h', '\r', c, 'i'})
		assert.NoError(t, err)
		if c == '\r' {
			assert.Equal(t, []byte("hi"), buf)
		} else {
			assert.Equal(t, []byte{'h', c, 'i'}, buf)
		}
	}
}

func TestSplitInput(t *testing.T) {
	var tests = []struct {
		in1      []byte
		len1     int
		in2      []byte
		len2     int
		expected []byte
	}{
		{[]byte{'h', IAC}, 1, []byte{NOP, 'i'}, 1, []byte("hi")},
		{[]byte{'h', IAC}, 1, []byte{IAC, 'i'}, 2, []byte{'h', IAC, 'i'}},
		{[]byte("foo\r"), 3, []byte("\nbar"), 4, []byte("foo\nbar")},
		{[]byte("foo\r"), 3, []byte("\x00bar"), 4, []byte("foo\rbar")},
	}
	for _, test := range tests {
		in := bytes.NewBuffer(test.in1)
		r := NewReader(in)
		buf := make([]byte, 16)
		n1, err := r.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, test.len1, n1)
		in.Write(test.in2)
		n2, err := r.Read(buf[n1:])
		assert.NoError(t, err)
		assert.Equal(t, test.len2, n2)
		assert.Equal(t, test.expected, buf[:n1+n2])
	}
}
