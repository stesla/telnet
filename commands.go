package telnet

import "fmt"

type commandByte byte

const (
	// RFC 885
	EOR = 239 + iota
	// RFC 854
	SE
	NOP
	DM
	BRK
	IP
	AO
	AYT
	EC
	EL
	GA
	SB
	WILL
	WONT
	DO
	DONT
	IAC
)

func (c commandByte) String() string {
	str, ok := map[commandByte]string{
		AO:   "AO",
		AYT:  "AYT",
		SB:   "SB",
		BRK:  "BRK",
		DM:   "DM",
		DO:   "DO",
		DONT: "DONT",
		SE:   "SE",
		EC:   "EC",
		EL:   "EL",
		GA:   "GA",
		IAC:  "IAC",
		IP:   "IP",
		NOP:  "NOP",
		WILL: "WILL",
		WONT: "WONT",
	}[c]
	if ok {
		return str
	}
	return fmt.Sprintf("%X", uint8(c))
}

type optionByte byte

const (
	TransmitBinary  = 0  // RFC 856
	Echo            = 1  // RFC 857
	SuppressGoAhead = 3  // RFC 858
	Charset         = 42 // RFC 2066
	TerminalType    = 24 // RFC 930
	NAWS            = 31 // RFC 1073
	EndOfRecord     = 25 // RFC 885
)

func (c optionByte) String() string {
	str, ok := map[optionByte]string{
		Charset:         "CHARSET",
		Echo:            "ECHO",
		EndOfRecord:     "END-OF-RECORD",
		NAWS:            "NAWS",
		SuppressGoAhead: "SUPPRESS-GO-AHEAD",
		TerminalType:    "TERMINAL-TYPE",
		TransmitBinary:  "TRANSMIT-BINARY",
	}[c]
	if ok {
		return str
	}
	return fmt.Sprintf("%X", uint8(c))
}

type charsetByte byte

const (
	charsetRequest = 1 + iota
	charsetAccepted
	charsetRejected
	charsetTTableIs
	charsetTTableRejected
	charsetTTableAck
	charsetTTableNak
)

func (c charsetByte) String() string {
	str, ok := map[charsetByte]string{
		charsetRequest:        "REQUEST",
		charsetAccepted:       "ACCEPTED",
		charsetRejected:       "REJECTED",
		charsetTTableIs:       "TTABLE-IS",
		charsetTTableRejected: "TTABLE-REJECTED",
		charsetTTableAck:      "TTABLE-ACK",
		charsetTTableNak:      "TTABLE-NAK",
	}[c]
	if ok {
		return str
	}
	return fmt.Sprintf("%X", uint8(c))
}

type telnetGoAhead struct{}

func (t telnetGoAhead) Error() string {
	return "IAC GA"
}

type telnetOptionCommand struct {
	cmd commandByte
	opt optionByte
}

func (t telnetOptionCommand) Error() string {
	return fmt.Sprintf("IAC %s %s", t.cmd, t.opt)
}

type telnetSubnegotiation struct {
	bytes []byte
}

func (t telnetSubnegotiation) Error() string {
	return fmt.Sprintf("IAC SB %q IAC SE", t.bytes)
}
