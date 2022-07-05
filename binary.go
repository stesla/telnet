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

type TransmitBinaryOption struct{}

func (t *TransmitBinaryOption) DisableForUs(conn Conn) {
	conn.SetWriteEncoding(ASCII)
}

func (t *TransmitBinaryOption) DisableForThem(conn Conn) {
	conn.SetReadEncoding(ASCII)
}

func (t *TransmitBinaryOption) EnableForUs(conn Conn) {
	conn.SetWriteEncoding(Binary)
}

func (t *TransmitBinaryOption) EnableForThem(conn Conn) {
	conn.SetReadEncoding(Binary)
}

func (t *TransmitBinaryOption) Option() byte { return TransmitBinary }

func (t *TransmitBinaryOption) Subnegotiation(_ Conn, _ []byte) {}
