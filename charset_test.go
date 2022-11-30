package telnet

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"golang.org/x/text/encoding/unicode"
)

func TestRejectIfNotEnabled(t *testing.T) {
	t.Skip("not implemented")
	var h OptionHandler = &CharsetOption{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	conn := NewMockConn(ctrl)
	h.DisableForUs(conn)
	h.DisableForThem(conn)

	expected := []byte{IAC, SB, Charset, charsetRejected, IAC, SE}
	conn.EXPECT().Send(expected)
	data := []byte{charsetRequest}
	data = append(data, "[TTABLE]\x01;US-ASCII;UTF-8"...)
	h.Subnegotiation(conn, data)
}

func TestAcceptASCII(t *testing.T) {
	var h OptionHandler = &CharsetOption{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	conn := NewMockConn(ctrl)
	h.EnableForUs(conn)

	expected := []byte{IAC, SB, Charset, charsetAccepted}
	expected = append(expected, "US-ASCII"...)
	expected = append(expected, IAC, SE)
	conn.EXPECT().Send(expected)
	conn.EXPECT().SetEncoding(ASCII)
	data := []byte{charsetRequest}
	data = append(data, "[TTABLE]\x01;ISO-8859-1;US-ASCII;CP437"...)
	h.Subnegotiation(conn, data)
}

func TestAcceptUTF8(t *testing.T) {
	var h OptionHandler = &CharsetOption{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	conn := NewMockConn(ctrl)
	h.EnableForUs(conn)

	expected := []byte{IAC, SB, Charset, charsetAccepted}
	expected = append(expected, "UTF-8"...)
	expected = append(expected, IAC, SE)
	conn.EXPECT().Send(expected)
	conn.EXPECT().SetEncoding(unicode.UTF8)
	data := []byte{charsetRequest}
	data = append(data, "[TTABLE]\x01;UTF-8;ISO-8859-1;US-ASCII;CP437"...)
	h.Subnegotiation(conn, data)
}
