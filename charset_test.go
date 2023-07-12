package telnet

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

func withCharsetAndConn(t *testing.T, f func(Option, *MockConn, *MockEventSink)) {
	var h Option = NewCharsetOption()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	conn := NewMockConn(ctrl)
	sink := NewMockEventSink(ctrl)
	conn.EXPECT().AddListener(h)
	h.Bind(conn, sink)
	assert.Equal(t, byte(Charset), h.Byte())
	f(h, conn, sink)
}

func expectRecvCharsetSubnegotiation(conn *MockConn, cmd charsetByte, v ...any) {
	args := []any{optionByte(Charset), cmd}
	args = append(args, v...)
	conn.EXPECT().Logf(
		"RECV: IAC SB %s %s %s IAC SE",
		args...,
	)
}

func expectCharsetRejected(conn *MockConn) {
	conn.EXPECT().Logf(
		"SEND: IAC SB %s %s IAC SE",
		optionByte(Charset),
		charsetByte(charsetRejected),
	)
}

func TestEmptySubnegotiationData(t *testing.T) {
	withCharsetAndConn(t, func(h Option, conn *MockConn, sink *MockEventSink) {
		conn.EXPECT().Logf(
			"RECV: IAC SB %s IAC SE",
			optionByte(Charset),
		)
		h.Subnegotiation([]byte{})
	})
}
func TestRejectIfNotEnabled(t *testing.T) {
	withCharsetAndConn(t, func(h Option, conn *MockConn, sink *MockEventSink) {
		expected := []byte{IAC, SB, Charset, charsetRejected, IAC, SE}
		conn.EXPECT().Send(expected)
		data := []byte{charsetRequest}
		subdata := []byte("[TTABLE]\x01;US-ASCII;UTF-8")
		data = append(data, subdata...)
		expectRecvCharsetSubnegotiation(conn, charsetRequest, string(subdata))
		expectCharsetRejected(conn)
		h.Subnegotiation(data)
	})
}

func TestRejectWhenEnabled(t *testing.T) {
	var tests = []string{
		"",
		";",
		"[TTABLE]\x01",
		"[TTABLE]\x01;",
		";BOGUS;ENCODING;NAMES",
	}
	for _, test := range tests {
		withCharsetAndConn(t, func(h Option, conn *MockConn, sink *MockEventSink) {
			expected := []byte{IAC, SB, Charset, charsetRejected, IAC, SE}
			conn.EXPECT().Send(expected)
			data := []byte{charsetRequest}
			data = append(data, test...)
			expectRecvCharsetSubnegotiation(conn, charsetRequest, test)
			expectCharsetRejected(conn)
			h.Subnegotiation(data)
		})
	}
}

func TestAcceptEncodingRequest(t *testing.T) {
	var tests = []struct {
		encoding             encoding.Encoding
		encodingName         string
		subnegotiationData   string
		binaryThem, binaryUs bool
		expected             bool
	}{
		{ASCII, "US-ASCII", "[TTABLE]\x01;US-ASCII;CP437", true, true, true},
		{charmap.ISO8859_1, "ISO-8859-1", ";ISO-8859-1;US-ASCII;CP437", true, true, true},
		{charmap.CodePage437, "CP437", ";CP437;US-ASCII", true, true, true},
		{unicode.UTF8, "UTF-8", ";UTF-8;ISO-8859-1;US-ASCII;CP437", true, true, true},
		{unicode.UTF8, "UTF-8", ";UTF-8;ISO-8859-1;US-ASCII;CP437", false, true, false},
		{unicode.UTF8, "UTF-8", ";UTF-8;ISO-8859-1;US-ASCII;CP437", true, false, false},
	}
	for _, test := range tests {
		withCharsetAndConn(t, func(h Option, conn *MockConn, sink *MockEventSink) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			co := h.(*CharsetOption)
			mockOption := NewMockOption(ctrl)
			mockOption.EXPECT().Conn().Return(conn).AnyTimes()
			mockOption.EXPECT().Sink().Return(sink).AnyTimes()
			mockOption.EXPECT().Byte().Return(byte(Charset)).AnyTimes()
			mockOption.EXPECT().EnabledForUs().Return(true).AnyTimes()
			co.Option = mockOption

			co.HandleEvent("update-option", UpdateOptionEvent{mockOption, false, true})
			expected := []byte{IAC, SB, Charset, charsetAccepted}
			expected = append(expected, test.encodingName...)
			expected = append(expected, IAC, SE)
			conn.EXPECT().Send(expected)

			mockBinary := NewMockOption(ctrl)
			mockBinary.EXPECT().Byte().Return(byte(TransmitBinary)).AnyTimes()
			mockBinary.EXPECT().EnabledForThem().Return(test.binaryThem).AnyTimes()
			mockBinary.EXPECT().EnabledForUs().Return(test.binaryUs).AnyTimes()
			conn.EXPECT().Option(uint8(TransmitBinary)).Return(mockBinary)
			if test.expected {
				conn.EXPECT().SetEncoding(test.encoding)
				sink.EXPECT().SendEvent("charset-accepted", test.encoding)
			} else {
				conn.EXPECT().SetEncoding(ASCII)
			}
			expectRecvCharsetSubnegotiation(conn, charsetRequest, test.subnegotiationData)
			conn.EXPECT().Logf(
				"SEND: IAC SB %s %s %s IAC SE",
				optionByte(Charset),
				charsetByte(charsetAccepted),
				test.encodingName,
			)
			data := []byte{charsetRequest}
			data = append(data, test.subnegotiationData...)
			h.Subnegotiation(data)

			assert.Equal(t, test.encoding, co.enc)
		})
	}
}

func TestEncodingRequestAccepted(t *testing.T) {
	withCharsetAndConn(t, func(h Option, conn *MockConn, sink *MockEventSink) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		co := h.(*CharsetOption)
		mockOption := NewMockOption(ctrl)
		mockOption.EXPECT().Conn().Return(conn).AnyTimes()
		mockOption.EXPECT().Sink().Return(sink).AnyTimes()
		mockOption.EXPECT().Byte().Return(byte(Charset)).AnyTimes()
		mockOption.EXPECT().EnabledForUs().Return(true).AnyTimes()
		co.Option = mockOption

		conn.EXPECT().Logf(
			"RECV: IAC SB %s %s %s IAC SE",
			optionByte(Charset),
			charsetByte(charsetAccepted),
			"UTF-8",
		)

		mockBinary := NewMockOption(ctrl)
		mockBinary.EXPECT().Byte().Return(byte(TransmitBinary)).AnyTimes()
		mockBinary.EXPECT().EnabledForThem().Return(true).AnyTimes()
		mockBinary.EXPECT().EnabledForUs().Return(true).AnyTimes()

		conn.EXPECT().Option(uint8(TransmitBinary)).Return(mockBinary)
		conn.EXPECT().SetEncoding(unicode.UTF8)
		sink.EXPECT().SendEvent("charset-accepted", unicode.UTF8)

		data := []byte{charsetAccepted}
		data = append(data, "UTF-8"...)
		h.Subnegotiation(data)
	})
}

func TestUpdateTransmitBinary(t *testing.T) {
	var tests = []struct {
		enabled           bool
		theyChanged, them bool
		weChanged, us     bool
		enc               encoding.Encoding
		expected          encoding.Encoding
	}{
		{true, false, false, false, false, nil, nil},
		{true, false, false, true, true, nil, nil},
		{true, true, true, false, false, nil, nil},
		{true, true, true, true, true, nil, nil},
		{true, false, true, false, true, unicode.UTF8, unicode.UTF8},
		{true, true, true, true, true, unicode.UTF8, unicode.UTF8},
		{true, true, false, true, true, unicode.UTF8, ASCII},
		{true, true, true, true, false, unicode.UTF8, ASCII},
		{false, true, true, true, true, unicode.UTF8, nil},
		{false, false, false, false, false, unicode.UTF8, nil},
	}
	for _, test := range tests {
		withCharsetAndConn(t, func(h Option, conn *MockConn, sink *MockEventSink) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			co := h.(*CharsetOption)
			co.enc = test.enc
			mockOption := NewMockOption(ctrl)
			mockOption.EXPECT().Conn().Return(conn).AnyTimes()
			mockOption.EXPECT().Sink().Return(sink).AnyTimes()
			mockOption.EXPECT().EnabledForUs().Return(test.enabled).AnyTimes()
			co.Option = mockOption

			if test.expected != nil {
				conn.EXPECT().SetEncoding(test.expected)
				if test.expected != ASCII {
					sink.EXPECT().SendEvent("charset-accepted", test.expected)
				}
			}

			mockBinary := NewMockOption(ctrl)
			mockBinary.EXPECT().Byte().Return(byte(TransmitBinary)).AnyTimes()
			mockBinary.EXPECT().EnabledForThem().Return(test.them).AnyTimes()
			mockBinary.EXPECT().EnabledForUs().Return(test.us).AnyTimes()

			co.HandleEvent("update-option", UpdateOptionEvent{mockBinary, test.theyChanged, test.weChanged})
		})
	}
}

func TestRejectsTTable(t *testing.T) {
	withCharsetAndConn(t, func(h Option, conn *MockConn, sink *MockEventSink) {
		conn.EXPECT().Logf(
			"RECV: IAC SB %s %s %s IAC SE",
			optionByte(Charset),
			charsetByte(charsetTTableIs),
			"\x01bogus",
		)
		conn.EXPECT().Send([]byte{IAC, SB, Charset, charsetTTableRejected, IAC, SE})

		data := []byte{charsetTTableIs, 1}
		data = append(data, "bogus"...)
		h.Subnegotiation(data)
	})
}

func TestCharsetRejected(t *testing.T) {
	withCharsetAndConn(t, func(h Option, conn *MockConn, sink *MockEventSink) {
		conn.EXPECT().Logf(
			"RECV: IAC SB %s %s %s IAC SE",
			optionByte(Charset),
			charsetByte(charsetRejected),
			"",
		)
		sink.EXPECT().SendEvent("charset-rejected", nil)

		data := []byte{charsetRejected}
		h.Subnegotiation(data)
	})
}
