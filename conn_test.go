package telnet

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/encoding/unicode"
)

func TestReadGoAhead(t *testing.T) {
	in := bytes.NewBuffer([]byte{'h', IAC, GA, 'i'})
	conn := newTestConn(in, nil)

	logger := NewMockLogger(t)
	conn.SetLogger(logger)

	logger.EXPECT().Logf("RECV: %s", []any{&telnetGoAhead{}})

	buf := make([]byte, 8)
	n, err := conn.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), buf[:n])
}

func TestWriteGoAhead(t *testing.T) {
	var out bytes.Buffer
	conn := newTestConn(nil, &out)
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
		"RECV: %s",
		[]any{&telnetOptionCommand{cmd, opt}},
	)
}

func expectSendOptionCommand(logger *MockLogger, cmd, opt byte) {
	logger.EXPECT().Logf(
		"SEND: IAC %s %s",
		[]any{
			commandByte(cmd),
			optionByte(opt),
		},
	)
}

func TestOption(t *testing.T) {
	var out bytes.Buffer
	in := bytes.NewBuffer([]byte{
		'h',
		IAC, WILL, Echo,
		IAC, DO, Echo,
		'i',
	})
	conn := newTestConn(in, &out)

	logger := NewMockLogger(t)
	conn.SetLogger(logger)

	expectReceiveOptionCommand(logger, WILL, Echo)
	expectSendOptionCommand(logger, DO, Echo)
	expectReceiveOptionCommand(logger, DO, Echo)
	expectSendOptionCommand(logger, WILL, Echo)

	option := NewOption(Echo)
	option.Allow(true, true)
	conn.BindOption(option)

	option2 := NewMockOption(t)
	option2.EXPECT().Bind(conn, conn)
	option2.EXPECT().Byte().Return(byte(TransmitBinary)).Maybe()
	conn.BindOption(option2)

	buf := make([]byte, 8)
	n, err := conn.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), buf[:n])
	assert.Equal(t, []byte{
		IAC, DO, Echo,
		IAC, WILL, Echo,
	}, out.Bytes())
	opt := conn.Option(Echo)
	assert.True(t, opt.EnabledForThem())
	assert.True(t, opt.EnabledForUs())

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
	n, err = conn.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), buf[:n])
	assert.Equal(t, []byte{
		IAC, DONT, Echo,
		IAC, WONT, Echo,
	}, out.Bytes())
	opt = conn.Option(Echo)
	assert.False(t, opt.EnabledForThem())
	assert.False(t, opt.EnabledForUs())
}

func TestEnableOption(t *testing.T) {
	conn := newTestConn(nil, nil)

	mockOption := NewMockOption(t)
	mockOption.EXPECT().Bind(conn, conn)
	mockOption.EXPECT().Byte().Return(byte(Echo)).Maybe()
	conn.BindOption(mockOption)

	mockOption.EXPECT().enableThem().Return(nil)
	conn.EnableOptionForThem(Echo, true)

	mockOption.EXPECT().enableUs().Return(nil)
	conn.EnableOptionForUs(Echo, true)

	mockOption.EXPECT().disableThem().Return(nil)
	conn.EnableOptionForThem(Echo, false)

	mockOption.EXPECT().disableUs().Return(nil)
	conn.EnableOptionForUs(Echo, false)
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
		conn := newTestConn(in, &out)
		buf, err := io.ReadAll(conn)
		assert.NoError(t, err, "test %d", i)
		assert.Equal(t, test.expectedr, buf, "test %d", i)
		assert.Equal(t, test.expectedw, out.Bytes(), "test %d", i)
	}
}

func TestASCIIByDefault(t *testing.T) {
	in := bytes.NewBuffer([]byte{'h', IAC, IAC, 'i'})
	var out bytes.Buffer
	conn := newTestConn(in, &out)
	conn.SuppressGoAhead(true)

	buf, err := io.ReadAll(conn)
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
	conn := newTestConn(in, &out)
	conn.SuppressGoAhead(true)
	conn.SetEncoding(unicode.UTF8)

	buf, err := io.ReadAll(conn)
	assert.NoError(t, err)
	assert.Equal(t, []byte("※"), buf)

	n, err := conn.Write([]byte("※"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte("※"), out.Bytes())
}

func TestSubnegotiation(t *testing.T) {
	in := bytes.NewBuffer([]byte{IAC, SB, Echo, 'h', 'i', IAC, SE})
	conn := newTestConn(in, nil)

	logger := NewMockLogger(t)
	conn.SetLogger(logger)

	option := NewMockOption(t)
	option.EXPECT().Byte().Return(byte(Echo)).Maybe()
	option.EXPECT().Bind(conn, conn)
	conn.BindOption(option)

	option.EXPECT().Subnegotiation([]byte("hi"))
	buf, err := io.ReadAll(conn)
	assert.NoError(t, err)
	assert.Empty(t, buf)
}

func TestSubnegotiationForUnsupportedOption(t *testing.T) {
	// This case should never actually happen, as subnegotiation should only
	// happen for options we've already negotiated. But, telnet implementations
	// don't always play by the rules, and if we're interacting with a broken
	// implementation, logging what they send us is good.
	in := bytes.NewBuffer([]byte{IAC, SB, Echo, 'h', 'i', IAC, SE})
	conn := newTestConn(in, nil)

	logger := NewMockLogger(t)
	logger.EXPECT().Logf(
		"RECV: IAC SB %s %q IAC SE",
		[]any{
			optionByte(Echo),
			[]byte("hi"),
		},
	)

	conn.SetLogger(logger)

	buf, err := io.ReadAll(conn)
	assert.NoError(t, err)
	assert.Empty(t, buf)
}

func TestSuppresGoAhead(t *testing.T) {
	h := NewSuppressGoAheadOption()
	assert.Implements(t, (*Option)(nil), h)

	conn := NewMockConn(t)

	conn.EXPECT().AddListener("update-option", h)
	h.Bind(conn, nil)

	assert.Equal(t, byte(SuppressGoAhead), h.Byte())

	opt := NewMockOption(t)
	opt.EXPECT().Byte().Return(byte(SuppressGoAhead)).Maybe()

	opt.EXPECT().EnabledForUs().Return(true).Once()
	conn.EXPECT().SuppressGoAhead(true)
	h.HandleEvent(UpdateOptionEvent{opt, false, true})

	opt.EXPECT().EnabledForUs().Return(false).Once()
	conn.EXPECT().SuppressGoAhead(false)
	h.HandleEvent(UpdateOptionEvent{opt, false, true})
}

func TestRequestCharset(t *testing.T) {
	var out bytes.Buffer
	conn := newTestConn(nil, &out)

	err := conn.RequestEncoding(unicode.UTF8)
	assert.Error(t, err)
	assert.Empty(t, out)

	charset := NewOption(Charset)
	charset.us = telnetQYes
	conn.BindOption(charset)

	listener := NewMockEventListener(t)
	conn.AddListener("charset-requested", listener)

	listener.EXPECT().HandleEvent(CharsetRequestedEvent{unicode.UTF8})
	err = conn.RequestEncoding(unicode.UTF8)
	assert.NoError(t, err)
	assert.Equal(t, []byte{IAC, SB, Charset, charsetRequest, ';', 'U', 'T', 'F', '-', '8', IAC, SE}, out.Bytes())
}

func TestSendEvent(t *testing.T) {
	conn := newTestConn(nil, nil)
	var called bool
	conn.AddListener("test-event", &FuncListener{func(event any) {
		called = true
		assert.Equal(t, "foo", event)
	}})
	conn.SendEvent("test-event", "foo")
	assert.True(t, called)
}

func TestRemoveListener(t *testing.T) {
	conn := newTestConn(nil, nil)
	count := 0
	fn := func(any) { count++ }
	listener := &FuncListener{fn}
	conn.AddListener("test-event", &FuncListener{fn})
	conn.AddListener("test-event", listener)
	conn.AddListener("test-event", &FuncListener{fn})
	conn.RemoveListener("test-event", listener)
	conn.SendEvent("test-event", "foo")
	assert.Equal(t, 2, count)
}

func TestDifferentEvents(t *testing.T) {
	conn := newTestConn(nil, nil)
	count := 0
	fn := func(any) { count++ }
	conn.AddListener("foo", &FuncListener{fn})
	conn.AddListener("baz", &FuncListener{fn})
	conn.SendEvent("foo", "bar")
	assert.Equal(t, 1, count)
}
