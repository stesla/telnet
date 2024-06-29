package telnet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUseBinary(t *testing.T) {
	var out bytes.Buffer
	conn := newTestConn(nil, &out)
	conn.SuppressGoAhead(true)
	conn.SetEncoding(Binary)
	n, err := conn.Write([]byte{'h', IAC, 'i'})
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte{'h', IAC, IAC, 'i'}, out.Bytes())
}

func TestTransmitBinaryOption(t *testing.T) {
	h := NewTransmitBinaryOption()
	assert.Implements(t, (*Option)(nil), h)

	conn := NewMockConn(t)
	conn.EXPECT().AddListener("update-option", h)
	h.Bind(conn, nil)

	assert.Equal(t, byte(TransmitBinary), h.Byte())

	opt := NewMockOption(t)
	opt.EXPECT().Byte().Return(byte(TransmitBinary)).Maybe()

	opt.EXPECT().EnabledForThem().Return(false).Once()
	opt.EXPECT().EnabledForUs().Return(false).Once()
	conn.EXPECT().SetReadEncoding(ASCII)
	conn.EXPECT().SetWriteEncoding(ASCII)
	h.HandleEvent(UpdateOptionEvent{opt, true, true})

	opt.EXPECT().EnabledForThem().Return(true).Once()
	opt.EXPECT().EnabledForUs().Return(true).Once()
	conn.EXPECT().SetReadEncoding(Binary)
	conn.EXPECT().SetWriteEncoding(Binary)
	h.HandleEvent(UpdateOptionEvent{opt, true, true})
}
