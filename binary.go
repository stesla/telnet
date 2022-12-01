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

func (t *TransmitBinaryOption) Option() byte { return TransmitBinary }

func (t *TransmitBinaryOption) Subnegotiation(_ Conn, _ []byte) {}

func (t *TransmitBinaryOption) Update(c Conn, option byte, theyChanged, them, weChanged, us bool) {
	if TransmitBinary != option {
		return
	}

	if theyChanged {
		if them {
			c.SetReadEncoding(Binary)
		} else {
			c.SetReadEncoding(ASCII)
		}
	}

	if weChanged {
		if us {
			c.SetWriteEncoding(Binary)
		} else {
			c.SetWriteEncoding(ASCII)
		}
	}
}
