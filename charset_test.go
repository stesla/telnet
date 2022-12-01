package telnet

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
)

func withCharsetAndConn(t *testing.T, f func(OptionHandler, *MockConn)) {
	var h OptionHandler = &CharsetOption{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	conn := NewMockConn(ctrl)
	f(h, conn)
}

func TestRejectIfNotEnabled(t *testing.T) {
	withCharsetAndConn(t, func(h OptionHandler, conn *MockConn) {
		h.DisableForUs(conn)

		expected := []byte{IAC, SB, Charset, charsetRejected, IAC, SE}
		conn.EXPECT().Send(expected)
		data := []byte{charsetRequest}
		data = append(data, "[TTABLE]\x01;US-ASCII;UTF-8"...)
		h.Subnegotiation(conn, data)
	})
}

func TestAcceptEncoding(t *testing.T) {
	var tests = []struct {
		encoding           encoding.Encoding
		encodingName       string
		subnegotiationData string
	}{
		{ASCII, "US-ASCII", "[TTABLE]\x01;ISO-8859-1;US-ASCII;CP437"},
		{unicode.UTF8, "UTF-8", "[TTABLE]\x01;UTF-8;ISO-8859-1;US-ASCII;CP437"},
	}
	for _, test := range tests {
		withCharsetAndConn(t, func(h OptionHandler, conn *MockConn) {
			h.EnableForUs(conn)
			expected := []byte{IAC, SB, Charset, charsetAccepted}
			expected = append(expected, test.encodingName...)
			expected = append(expected, IAC, SE)
			conn.EXPECT().Send(expected)
			conn.EXPECT().SetEncoding(test.encoding)
			data := []byte{charsetRequest}
			data = append(data, test.subnegotiationData...)
			h.Subnegotiation(conn, data)
		})
	}
}
