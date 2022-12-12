package telnet

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

func withCharsetAndConn(t *testing.T, f func(OptionHandler, *MockConn)) {
	var h OptionHandler = &CharsetOption{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	conn := NewMockConn(ctrl)
	f(h, conn)
}

func expectCharsetRejected(conn *MockConn) {
	conn.EXPECT().Logf(
		DEBUG,
		"SEND: IAC SB %s %s IAC SE",
		optionByte(Charset),
		charsetByte(charsetRejected),
	)
}

func TestRejectIfNotEnabled(t *testing.T) {
	withCharsetAndConn(t, func(h OptionHandler, conn *MockConn) {
		expected := []byte{IAC, SB, Charset, charsetRejected, IAC, SE}
		conn.EXPECT().Send(expected)
		data := []byte{charsetRequest}
		subdata := []byte("[TTABLE]\x01;US-ASCII;UTF-8")
		data = append(data, subdata...)
		expectCharsetRejected(conn)
		h.Subnegotiation(conn, data)
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
		withCharsetAndConn(t, func(h OptionHandler, conn *MockConn) {
			h.Update(conn, uint8(Charset), false, false, true, true)
			expected := []byte{IAC, SB, Charset, charsetRejected, IAC, SE}
			conn.EXPECT().Send(expected)
			data := []byte{charsetRequest}
			data = append(data, test...)
			expectCharsetRejected(conn)
			h.Subnegotiation(conn, data)
		})
	}
}

func TestAcceptEncoding(t *testing.T) {
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
		withCharsetAndConn(t, func(h OptionHandler, conn *MockConn) {
			h.Update(conn, uint8(Charset), false, false, true, true)
			expected := []byte{IAC, SB, Charset, charsetAccepted}
			expected = append(expected, test.encodingName...)
			expected = append(expected, IAC, SE)
			conn.EXPECT().Send(expected)
			conn.EXPECT().OptionEnabled(uint8(TransmitBinary)).Return(test.binaryThem, test.binaryUs)
			if test.expected {
				conn.EXPECT().SetEncoding(test.encoding)
			} else {
				conn.EXPECT().SetEncoding(ASCII)
			}
			conn.EXPECT().Logf(
				DEBUG,
				"SEND: IAC SB %s %s %s IAC SE",
				optionByte(Charset),
				charsetByte(charsetAccepted),
				test.encodingName,
			)
			data := []byte{charsetRequest}
			data = append(data, test.subnegotiationData...)
			h.Subnegotiation(conn, data)

			co := h.(*CharsetOption)
			assert.Equal(t, test.encoding, co.enc)
		})
	}
}

func TestUpdateCharset(t *testing.T) {
	var tests = []struct {
		theyChanged, them                      bool
		weChanged, us                          bool
		expectedThemEnabled, expectedUsEnabled bool
	}{
		{false, false, false, false, false, false},
		{false, true, false, true, true, true},
		{true, false, true, false, false, false},
		{true, false, true, true, false, true},
	}
	for _, test := range tests {
		withCharsetAndConn(t, func(h OptionHandler, conn *MockConn) {
			h.Update(conn, Charset, test.theyChanged, test.them, test.weChanged, test.us)
			co := h.(*CharsetOption)
			assert.Equal(t, test.expectedThemEnabled, co.enabledForThem)
			assert.Equal(t, test.expectedUsEnabled, co.enabledForUs)
		})
	}
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
		withCharsetAndConn(t, func(h OptionHandler, conn *MockConn) {
			if test.enabled {
				h.Update(conn, uint8(Charset), false, false, true, true)
			}

			co := h.(*CharsetOption)
			co.enc = test.enc

			if test.expected != nil {
				conn.EXPECT().SetEncoding(test.expected)
			}
			h.Update(conn, TransmitBinary, test.theyChanged, test.them, test.weChanged, test.us)
		})
	}
}
