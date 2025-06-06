package telnet

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

func withCharsetAndConn(t *testing.T, f func(*CharsetOption, *MockConn, *MockEventSink)) {
	h := NewCharsetOption(true, false)
	assert.Implements(t, (*Option)(nil), h)
	conn := NewMockConn(t)
	sink := NewMockEventSink(t)
	conn.EXPECT().AddListener("update-option", h)
	h.Bind(conn, sink)
	assert.Equal(t, byte(Charset), h.Byte())
	f(h, conn, sink)
}

func expectRecvCharsetSubnegotiation(conn *MockConn, cmd charsetByte, v ...any) {
	args := []any{optionByte(Charset), cmd}
	args = append(args, v...)
	conn.EXPECT().Logf(
		"RECV: IAC SB %s %s %s IAC SE",
		args,
	)
}

func expectCharsetRejected(conn *MockConn) {
	conn.EXPECT().Logf(
		"SEND: IAC SB %s %s IAC SE",
		[]any{
			optionByte(Charset),
			charsetByte(charsetRejected),
		},
	)
}

func TestEmptySubnegotiationData(t *testing.T) {
	withCharsetAndConn(t, func(h *CharsetOption, conn *MockConn, sink *MockEventSink) {
		conn.EXPECT().Logf(
			"RECV: IAC SB %s IAC SE",
			[]any{optionByte(Charset)},
		)
		h.Subnegotiation([]byte{})
	})
}
func TestRejectIfNotEnabled(t *testing.T) {
	withCharsetAndConn(t, func(h *CharsetOption, conn *MockConn, sink *MockEventSink) {
		expected := []byte{IAC, SB, Charset, charsetRejected, IAC, SE}
		conn.EXPECT().Send(expected).Return(len(expected), nil)
		data := []byte{charsetRequest}
		subdata := []byte("[TTABLE]\x01;US-ASCII;UTF-8")
		data = append(data, subdata...)
		expectRecvCharsetSubnegotiation(conn, charsetRequest, string(subdata))
		expectCharsetRejected(conn)
		h.Subnegotiation(data)
	})
}

func TestRejectIfServerAlreadySentRequest(t *testing.T) {
	withCharsetAndConn(t, func(h *CharsetOption, conn *MockConn, sink *MockEventSink) {
		option := NewMockOption(t)
		option.EXPECT().Byte().Return(byte(Charset)).Maybe()
		option.EXPECT().EnabledForUs().Return(true).Maybe()
		h.isServer = true
		h.HandleEvent(CharsetRequestedEvent{unicode.UTF8})
		expected := []byte{IAC, SB, Charset, charsetRejected, IAC, SE}
		conn.EXPECT().Send(expected).Return(len(expected), nil)
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
		withCharsetAndConn(t, func(h *CharsetOption, conn *MockConn, sink *MockEventSink) {
			expected := []byte{IAC, SB, Charset, charsetRejected, IAC, SE}
			conn.EXPECT().Send(expected).Return(len(expected), nil)
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
		requireBinary        bool
	}{
		{ASCII, "US-ASCII", "[TTABLE]\x01;US-ASCII;CP437", true, true, true, true},
		{charmap.ISO8859_1, "ISO-8859-1", ";ISO-8859-1;US-ASCII;CP437", true, true, true, true},
		{charmap.CodePage437, "CP437", ";CP437;US-ASCII", true, true, true, true},
		{unicode.UTF8, "UTF-8", ";UTF-8;ISO-8859-1;US-ASCII;CP437", true, true, true, true},
		{unicode.UTF8, "UTF-8", ";UTF-8;ISO-8859-1;US-ASCII;CP437", false, true, false, true},
		{unicode.UTF8, "UTF-8", ";UTF-8;ISO-8859-1;US-ASCII;CP437", true, false, false, true},
		{unicode.UTF8, "UTF-8", ";UTF-8", false, false, true, false},
	}
	for _, test := range tests {
		withCharsetAndConn(t, func(h *CharsetOption, conn *MockConn, sink *MockEventSink) {
			h.requireBinary = test.requireBinary

			mockOption := NewMockOption(t)
			mockOption.EXPECT().Conn().Return(conn).Maybe()
			mockOption.EXPECT().Sink().Return(sink).Maybe()
			mockOption.EXPECT().Byte().Return(byte(Charset)).Maybe()
			mockOption.EXPECT().EnabledForUs().Return(true).Maybe()
			h.Option = mockOption

			h.HandleEvent(UpdateOptionEvent{mockOption, false, true})
			expected := []byte{IAC, SB, Charset, charsetAccepted}
			expected = append(expected, test.encodingName...)
			expected = append(expected, IAC, SE)
			conn.EXPECT().Send(expected).Return(len(expected), nil)

			mockBinary := NewMockOption(t)
			mockBinary.EXPECT().Byte().Return(byte(TransmitBinary)).Maybe()
			mockBinary.EXPECT().EnabledForThem().Return(test.binaryThem).Maybe()
			mockBinary.EXPECT().EnabledForUs().Return(test.binaryUs).Maybe()
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
				[]any{
					optionByte(Charset),
					charsetByte(charsetAccepted),
					test.encodingName,
				},
			)
			data := []byte{charsetRequest}
			data = append(data, test.subnegotiationData...)

			h.requestedEnc = unicode.UTF8
			h.Subnegotiation(data)
			assert.Equal(t, test.encoding, h.enc)
		})
	}
}

func TestEncodingRequestAccepted(t *testing.T) {
	withCharsetAndConn(t, func(h *CharsetOption, conn *MockConn, sink *MockEventSink) {
		mockOption := NewMockOption(t)
		mockOption.EXPECT().Conn().Return(conn).Maybe()
		mockOption.EXPECT().Sink().Return(sink).Maybe()
		mockOption.EXPECT().Byte().Return(byte(Charset)).Maybe()
		mockOption.EXPECT().EnabledForUs().Return(true).Maybe()
		h.Option = mockOption

		conn.EXPECT().Logf(
			"RECV: IAC SB %s %s %s IAC SE",
			[]any{
				optionByte(Charset),
				charsetByte(charsetAccepted),
				"UTF-8",
			},
		)

		mockBinary := NewMockOption(t)
		mockBinary.EXPECT().Byte().Return(byte(TransmitBinary)).Maybe()
		mockBinary.EXPECT().EnabledForThem().Return(true).Maybe()
		mockBinary.EXPECT().EnabledForUs().Return(true).Maybe()

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
		withCharsetAndConn(t, func(h *CharsetOption, conn *MockConn, sink *MockEventSink) {
			h.enc = test.enc
			mockOption := NewMockOption(t)
			mockOption.EXPECT().Conn().Return(conn).Maybe()
			mockOption.EXPECT().Sink().Return(sink).Maybe()
			mockOption.EXPECT().EnabledForUs().Return(test.enabled).Maybe()
			h.Option = mockOption

			if test.expected != nil {
				conn.EXPECT().SetEncoding(test.expected)
				if test.expected != ASCII {
					sink.EXPECT().SendEvent("charset-accepted", test.expected)
				}
			}

			mockBinary := NewMockOption(t)
			mockBinary.EXPECT().Byte().Return(byte(TransmitBinary)).Maybe()
			mockBinary.EXPECT().EnabledForThem().Return(test.them).Maybe()
			mockBinary.EXPECT().EnabledForUs().Return(test.us).Maybe()

			h.HandleEvent(UpdateOptionEvent{mockBinary, test.theyChanged, test.weChanged})
		})
	}
}

func TestRejectsTTable(t *testing.T) {
	withCharsetAndConn(t, func(h *CharsetOption, conn *MockConn, sink *MockEventSink) {
		conn.EXPECT().Logf(
			"RECV: IAC SB %s %s %s IAC SE",
			[]any{
				optionByte(Charset),
				charsetByte(charsetTTableIs),
				"\x01bogus",
			},
		)
		expected := []byte{IAC, SB, Charset, charsetTTableRejected, IAC, SE}
		conn.EXPECT().Send(expected).Return(len(expected), nil)

		data := []byte{charsetTTableIs, 1}
		data = append(data, "bogus"...)
		h.Subnegotiation(data)
	})
}

func TestCharsetRejected(t *testing.T) {
	withCharsetAndConn(t, func(h *CharsetOption, conn *MockConn, sink *MockEventSink) {
		conn.EXPECT().Logf(
			"RECV: IAC SB %s %s %s IAC SE",
			[]any{
				optionByte(Charset),
				charsetByte(charsetRejected),
				"",
			},
		)
		sink.EXPECT().SendEvent("charset-rejected", nil)

		data := []byte{charsetRejected}
		h.Subnegotiation(data)
	})
}
