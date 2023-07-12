package telnet

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

var Binary encoding.Encoding = &binaryEncoding{}

type binaryEncoding struct{}

func (e binaryEncoding) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{Transformer: e}
}

func (e binaryEncoding) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{Transformer: e}
}

func (binaryEncoding) String() string { return "ASCII" }

func (e binaryEncoding) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for i, c := range src {
		if nDst >= len(dst) {
			err = transform.ErrShortDst
			break
		}
		dst[nDst] = c
		nDst++
		nSrc = i + 1
	}
	return
}

func (a binaryEncoding) Reset() {}

type TransmitBinaryOption struct {
	Option
}

func NewTransmitBinaryOption() *TransmitBinaryOption {
	return &TransmitBinaryOption{Option: NewOption(TransmitBinary)}
}

func (t *TransmitBinaryOption) Bind(conn Conn, sink EventSink) {
	t.Option.Bind(conn, sink)
	conn.AddListener(t)
}

func (t *TransmitBinaryOption) Subnegotiation(_ []byte) {}

func (t *TransmitBinaryOption) HandleEvent(name string, data any) {
	event, ok := data.(UpdateOptionEvent)
	if !ok {
		return
	}

	if event.TheyChanged {
		if event.Option.EnabledForThem() {
			t.Conn().SetReadEncoding(Binary)
		} else {
			t.Conn().SetReadEncoding(ASCII)
		}
	}

	if event.WeChanged {
		if event.Option.EnabledForUs() {
			t.Conn().SetWriteEncoding(Binary)
		} else {
			t.Conn().SetWriteEncoding(ASCII)
		}
	}
}
