package telnet

import (
	"bytes"
	"io/ioutil"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/encoding/unicode"
)

func TestReadGoAhead(t *testing.T) {
	in := bytes.NewBuffer([]byte{'h', IAC, GA, 'i'})
	conn := newConnection(in, nil)
	buf := make([]byte, 8)
	n, err := conn.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), buf[:n])
}

func TestWriteGoAhead(t *testing.T) {
	var out bytes.Buffer
	conn := newConnection(nil, &out)
	n, err := conn.Write([]byte("foo"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte{'f', 'o', 'o', IAC, GA}, out.Bytes())
	out.Reset()

	conn.SuppressGoAhead(true)
	n, err = conn.Write([]byte("foo"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte("foo"), out.Bytes())
}

func TestOptionHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var out bytes.Buffer
	in := bytes.NewBuffer([]byte{
		'h',
		IAC, WILL, Echo,
		IAC, DO, Echo,
		'i',
	})
	conn := newConnection(in, &out)

	handler := NewMockOptionHandler(ctrl)
	handler.EXPECT().Option().Return(byte(Echo))
	conn.AllowOption(handler, true, true)
	opt := conn.opts.get(Echo)

	buf := make([]byte, 8)
	handler.EXPECT().Enable(conn)
	n, err := conn.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), buf[:n])
	assert.Equal(t, []byte{
		IAC, DO, Echo,
		IAC, WILL, Echo,
	}, out.Bytes())
	assert.True(t, opt.enabledForThem())
	assert.True(t, opt.enabledForUs())

	buf = make([]byte, 8)
	in.Write([]byte{
		'h',
		IAC, WONT, Echo,
		IAC, DONT, Echo,
		'i',
	})
	out.Reset()
	handler.EXPECT().Disable(conn)
	n, err = conn.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), buf[:n])
	assert.Equal(t, []byte{
		IAC, DONT, Echo,
		IAC, WONT, Echo,
	}, out.Bytes())
	assert.False(t, opt.enabledForThem())
	assert.False(t, opt.enabledForUs())
}

func TestEnableOption(t *testing.T) {
	var out bytes.Buffer
	conn := newConnection(nil, &out)

	conn.EnableOptionForThem(Echo, true)
	conn.EnableOptionForUs(Echo, true)
	assert.Equal(t, []byte{
		IAC, DO, Echo,
		IAC, WILL, Echo,
	}, out.Bytes())
	out.Reset()

	opt := conn.opts.get(Echo)
	opt.them = telnetQYes
	opt.us = telnetQYes

	conn.EnableOptionForThem(Echo, false)
	conn.EnableOptionForUs(Echo, false)
	assert.Equal(t, []byte{
		IAC, DONT, Echo,
		IAC, WONT, Echo,
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

func TestASCIIByDefault(t *testing.T) {
	in := bytes.NewBuffer([]byte{'h', IAC, IAC, 'i'})
	var out bytes.Buffer
	conn := newConnection(in, &out)
	conn.SuppressGoAhead(true)

	buf, err := ioutil.ReadAll(conn)
	assert.NoError(t, err)
	assert.Equal(t, []byte("h\x1ai"), buf)
	out.Reset()

	n, err := conn.Write([]byte{'h', IAC, 'i'})
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte("h\x1ai"), out.Bytes())
}

func TestSetReadEncoding(t *testing.T) {
	in := bytes.NewBuffer([]byte{0xe2, 0x80, 0xbb})
	var out bytes.Buffer
	conn := newConnection(in, &out)
	conn.SuppressGoAhead(true)
	conn.SetEncoding(unicode.UTF8)

	buf, err := ioutil.ReadAll(conn)
	assert.NoError(t, err)
	assert.Equal(t, []byte("※"), buf)

	n, err := conn.Write([]byte("※"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte("※"), out.Bytes())
}

func TestSubnegotiation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	in := bytes.NewBuffer([]byte{IAC, SB, Echo, 'h', 'i', IAC, SE})
	conn := newConnection(in, nil)

	handler := NewMockOptionHandler(ctrl)
	handler.EXPECT().Option().Return(byte(Echo))
	conn.AllowOption(handler, true, true)
	opt := conn.opts.get(Echo)
	opt.them, opt.us = telnetQYes, telnetQYes

	handler.EXPECT().Subnegotiation(conn, []byte("hi"))
	buf, err := ioutil.ReadAll(conn)
	assert.NoError(t, err)
	assert.Empty(t, buf)
}
