package telnet

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadGoAhead(t *testing.T) {
	in := bytes.NewBuffer([]byte{'h', IAC, GA, 'i'})
	conn := newConnection(in, nil)
	buf := make([]byte, 8)
	n1, err := conn.Read(buf)
	assert.NoError(t, err)
	n2, err := conn.Read(buf[n1:])
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), buf[:n1+n2])
}

func TestWriteGoAhead(t *testing.T) {
	var out bytes.Buffer
	conn := newConnection(nil, &out)
	n, err := conn.Write([]byte("foo"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte{'f', 'o', 'o', IAC, GA}, out.Bytes())
	out.Reset()

	conn.opts.get(SuppressGoAhead).us = telnetQYes
	n, err = conn.Write([]byte("foo"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte("foo"), out.Bytes())
}

func TestAllowOption(t *testing.T) {
	var out bytes.Buffer
	in := bytes.NewBuffer([]byte{
		IAC, WILL, SuppressGoAhead,
		IAC, DO, SuppressGoAhead,
	})
	conn := newConnection(in, &out)
	conn.AllowOptionForThem(SuppressGoAhead, true)
	conn.AllowOptionForUs(SuppressGoAhead, true)

	buf, err := ioutil.ReadAll(conn)
	assert.NoError(t, err)
	assert.Empty(t, buf)
	assert.Equal(t, []byte{
		IAC, DO, SuppressGoAhead,
		IAC, WILL, SuppressGoAhead,
	}, out.Bytes())

	opt := conn.opts.get(SuppressGoAhead)
	assert.True(t, opt.enabledForThem())
	assert.True(t, opt.enabledForUs())
}

func TestEnableOption(t *testing.T) {
	var out bytes.Buffer
	conn := newConnection(nil, &out)

	conn.EnableOptionForThem(SuppressGoAhead, true)
	conn.EnableOptionForUs(SuppressGoAhead, true)
	assert.Equal(t, []byte{
		IAC, DO, SuppressGoAhead,
		IAC, WILL, SuppressGoAhead,
	}, out.Bytes())
	out.Reset()

	opt := conn.opts.get(SuppressGoAhead)
	opt.them = telnetQYes
	opt.us = telnetQYes

	conn.EnableOptionForThem(SuppressGoAhead, false)
	conn.EnableOptionForUs(SuppressGoAhead, false)
	assert.Equal(t, []byte{
		IAC, DONT, SuppressGoAhead,
		IAC, WONT, SuppressGoAhead,
	}, out.Bytes())
	out.Reset()
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
		buf, err := ioutil.ReadAll(conn)
		assert.NoError(t, err, "test %d", i)
		assert.Equal(t, test.expectedr, buf, "test %d", i)
		assert.Equal(t, test.expectedw, out.Bytes(), "test %d", i)
	}
}
