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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	in := bytes.NewBuffer([]byte{'h', IAC, GA, 'i'})
	conn := newConnection(in, nil)

	logger := NewMockLogger(ctrl)
	conn.SetLogger(logger)

	logger.EXPECT().Logf(DEBUG, "RECV: %s", &telnetGoAhead{})

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

func expectReceiveOptionCommand(logger *MockLogger, cmd, opt byte) {
	logger.EXPECT().Logf(
		DEBUG,
		"RECV: %s",
		&telnetOptionCommand{cmd, opt},
	)
}

func expectSendOptionCommand(logger *MockLogger, cmd, opt byte) {
	logger.EXPECT().Logf(
		DEBUG,
		"SEND: IAC %s %s",
		commandByte(cmd),
		optionByte(opt),
	)
}

func TestOption(t *testing.T) {
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

	logger := NewMockLogger(ctrl)
	conn.SetLogger(logger)

	expectReceiveOptionCommand(logger, WILL, Echo)
	expectSendOptionCommand(logger, DO, Echo)
	expectReceiveOptionCommand(logger, DO, Echo)
	expectSendOptionCommand(logger, WILL, Echo)

	option := NewOption(Echo)
	option.Allow(true, true)
	conn.BindOption(option)

	option2 := NewMockOption(ctrl)
	option2.EXPECT().Bind(conn)
	option2.EXPECT().Byte().Return(byte(TransmitBinary)).AnyTimes()
	conn.BindOption(option2)

	buf := make([]byte, 8)
	option2.EXPECT().Update(uint8(Echo), true, true, false, false)
	option2.EXPECT().Update(uint8(Echo), false, true, true, true)
	n, err := conn.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), buf[:n])
	assert.Equal(t, []byte{
		IAC, DO, Echo,
		IAC, WILL, Echo,
	}, out.Bytes())
	them, us := conn.OptionEnabled(Echo)
	assert.True(t, them)
	assert.True(t, us)

	expectReceiveOptionCommand(logger, WONT, Echo)
	expectSendOptionCommand(logger, DONT, Echo)
	expectReceiveOptionCommand(logger, DONT, Echo)
	expectSendOptionCommand(logger, WONT, Echo)

	buf = make([]byte, 8)
	in.Write([]byte{
		'h',
		IAC, WONT, Echo,
		IAC, DONT, Echo,
		'i',
	})
	out.Reset()
	option2.EXPECT().Update(uint8(Echo), true, false, false, true)
	option2.EXPECT().Update(uint8(Echo), false, false, true, false)
	n, err = conn.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), buf[:n])
	assert.Equal(t, []byte{
		IAC, DONT, Echo,
		IAC, WONT, Echo,
	}, out.Bytes())
	them, us = conn.OptionEnabled(Echo)
	assert.False(t, them)
	assert.False(t, us)
}

func TestEnableOption(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn := newConnection(nil, nil)

	mockOption := NewMockOption(ctrl)
	mockOption.EXPECT().Bind(conn)
	mockOption.EXPECT().Byte().Return(byte(Echo)).AnyTimes()
	conn.BindOption(mockOption)

	mockOption.EXPECT().enableThem(conn)
	conn.EnableOptionForThem(Echo, true)

	mockOption.EXPECT().enableUs(conn)
	conn.EnableOptionForUs(Echo, true)

	mockOption.EXPECT().disableThem(conn)
	conn.EnableOptionForThem(Echo, false)

	mockOption.EXPECT().disableUs(conn)
	conn.EnableOptionForUs(Echo, false)
}

func TestSendOptionCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var out bytes.Buffer
	conn := newConnection(nil, &out)

	logger := NewMockLogger(ctrl)
	conn.SetLogger(logger)

	expectSendOptionCommand(logger, WONT, Echo)
	conn.sendOptionCommand(WONT, Echo)
	assert.Equal(t, []byte{IAC, WONT, Echo}, out.Bytes())
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

	logger := NewMockLogger(ctrl)
	conn.SetLogger(logger)

	option := NewMockOption(ctrl)
	option.EXPECT().Byte().Return(byte(Echo)).AnyTimes()
	option.EXPECT().Bind(conn)
	conn.BindOption(option)

	option.EXPECT().Subnegotiation([]byte("hi"))
	buf, err := ioutil.ReadAll(conn)
	assert.NoError(t, err)
	assert.Empty(t, buf)
}

func TestSubnegotiationForUnsupportedOption(t *testing.T) {
	// This case should never actually happen, as subnegotiation should only
	// happen for options we've already negotiated. But, telnet implementations
	// don't always play by the rules, and if we're interacting with a broken
	// implementation, logging what they send us is good.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	in := bytes.NewBuffer([]byte{IAC, SB, Echo, 'h', 'i', IAC, SE})
	conn := newConnection(in, nil)

	logger := NewMockLogger(ctrl)
	logger.EXPECT().Logf(
		DEBUG,
		"RECV: IAC SB %s %q IAC SE",
		optionByte(Echo),
		[]byte("hi"),
	)

	conn.SetLogger(logger)

	buf, err := ioutil.ReadAll(conn)
	assert.NoError(t, err)
	assert.Empty(t, buf)
}

func TestSuppresGoAhead(t *testing.T) {
	var h Option = NewSuppressGoAheadOption()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	conn := NewMockConn(ctrl)
	h.Bind(conn)

	assert.Equal(t, byte(SuppressGoAhead), h.Byte())

	conn.EXPECT().SuppressGoAhead(true)
	h.Update(uint8(SuppressGoAhead), false, false, true, true)

	conn.EXPECT().SuppressGoAhead(false)
	h.Update(uint8(SuppressGoAhead), true, true, true, false)

	h.Update(uint8(Echo), true, true, true, true)
}
