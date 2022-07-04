package telnet

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readTest(in []byte) (out []byte, err error) {
	r := NewReader(bytes.NewBuffer(in), nil)
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
		{[]byte{'h', IAC, SB, IAC, SE, 'i'}, []byte("hi")},
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
		r := NewReader(in, nil)
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

type boomReader int

func (r boomReader) Read(b []byte) (n int, err error) {
	for i := 0; i < int(r) && i < len(b); i++ {
		b[i] = 'A' + byte(i)
	}
	return int(r), errors.New("boom")
}

func TestErrorReading(t *testing.T) {
	r := NewReader(boomReader(3), nil)
	buf := make([]byte, 16)
	n, err := r.Read(buf)
	buf = buf[:n]
	assert.ErrorContains(t, err, "boom")
	assert.Equal(t, []byte("ABC"), buf)
}

func TestEOFOnSeparateRead(t *testing.T) {
	r := NewReader(bytes.NewBufferString("hi"), nil)
	buf := make([]byte, 16)
	n, err := r.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), buf[:n])
	n, err = r.Read(buf)
	assert.Error(t, io.EOF)
	assert.Equal(t, 0, n)
}

func TestNilCommandHandler(t *testing.T) {
	r := NewReader(bytes.NewBuffer([]byte{'h', IAC, GA, 'i'}), nil)
	buf := make([]byte, 16)
	n, err := r.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), buf[:n])
}

func TestErrorInCommandHandler(t *testing.T) {
	r := NewReader(bytes.NewBuffer([]byte{'h', IAC, GA, 'i'}), func(any) error {
		return errors.New("boom")
	})
	buf := make([]byte, 16)
	n, err := r.Read(buf)
	assert.Error(t, err)
	assert.Equal(t, []byte("h"), buf[:n])
}

func TestTelnetCommand(t *testing.T) {
	var tests = []struct {
		in, expected []byte
		cmd          fmt.Stringer
	}{
		{[]byte{'h', IAC, GA, 'i'}, []byte("hi"), &telnetGoAhead{}},
		{[]byte{'h', IAC, DO, Echo, 'i'}, []byte("hi"), &telnetOptionCommand{DO, Echo}},
		{[]byte{'h', IAC, DONT, Echo, 'i'}, []byte("hi"), &telnetOptionCommand{DONT, Echo}},
		{[]byte{'h', IAC, WILL, Echo, 'i'}, []byte("hi"), &telnetOptionCommand{WILL, Echo}},
		{[]byte{'h', IAC, WONT, Echo, 'i'}, []byte("hi"), &telnetOptionCommand{WONT, Echo}},
		{[]byte{'h', IAC, SB, 'f', 'o', 'o', IAC, SE, 'i'}, []byte("hi"), &telnetSubnegotiation{[]byte("foo")}},
		{[]byte{'h', IAC, SB, IAC, IAC, IAC, SE, 'i'}, []byte("hi"), &telnetSubnegotiation{[]byte{IAC}}},
	}
	for _, test := range tests {
		var actual any
		r := NewReader(bytes.NewBuffer(test.in), func(cmd any) error {
			actual = cmd
			return nil
		})
		buf := make([]byte, 16)
		n, err := r.Read(buf)
		assert.NoError(t, err, test.cmd.String())
		assert.Equal(t, test.expected, buf[:n], test.cmd.String())
		assert.Equal(t, test.cmd, actual, test.cmd.String())

		r = NewReader(bytes.NewBuffer(test.in), func(cmd any) error {
			return errors.New("boom")
		})
		n, err = r.Read(buf)
		assert.Error(t, err, test.cmd.String())
		assert.Equal(t, test.expected[:1], buf[:n], test.cmd.String())
	}
}
