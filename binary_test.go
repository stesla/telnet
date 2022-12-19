package telnet

import (
	"bytes"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUseBinary(t *testing.T) {
	var out bytes.Buffer
	conn := newConnection(nil, &out)
	conn.SuppressGoAhead(true)
	conn.SetEncoding(Binary)
	n, err := conn.Write([]byte{'h', IAC, 'i'})
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte{'h', IAC, IAC, 'i'}, out.Bytes())
}

func TestTransmitBinaryOption(t *testing.T) {
	var h Option = NewTransmitBinaryOption()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	conn := NewMockConn(ctrl)

	assert.Equal(t, byte(TransmitBinary), h.Byte())

	conn.EXPECT().SetReadEncoding(ASCII)
	conn.EXPECT().SetWriteEncoding(ASCII)
	h.Update(conn, uint8(TransmitBinary), true, false, true, false)

	conn.EXPECT().SetReadEncoding(Binary)
	conn.EXPECT().SetWriteEncoding(Binary)
	h.Update(conn, uint8(TransmitBinary), true, true, true, true)
}
