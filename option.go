package telnet

import "math"

type Option interface {
	Allow(them, us bool)
	Bind(Conn)
	Byte() byte
	Conn() Conn
	EnabledForThem() bool
	EnabledForUs() bool
	Subnegotiation([]byte)
	Update(option byte, theyChanged, them, weChanged, us bool)

	disableThem() error
	disableUs() error
	enableThem() error
	enableUs() error
	receive(c byte) error
}

func newOptionMap() *optionMap {
	m := make(map[byte]Option)
	for b := byte(0); b < math.MaxUint8; b++ {
		m[b] = NewOption(b)
	}
	return &optionMap{m: m}
}

type optionMap struct {
	m map[byte]Option
}

func (m *optionMap) each(fn func(Option)) {
	for _, opt := range m.m {
		fn(opt)
	}
}

func (m *optionMap) get(c byte) Option {
	return m.m[c]
}

func (m *optionMap) put(o Option) {
	m.m[o.Byte()] = o
}

type option struct {
	conn               Conn
	code               byte
	allowUs, allowThem bool
	us, them           telnetQState
}

func NewOption(c byte) *option {
	return &option{code: c}
}

func (o *option) Allow(them, us bool)  { o.allowThem, o.allowUs = them, us }
func (o *option) Bind(conn Conn)       { o.conn = conn }
func (o *option) Byte() byte           { return o.code }
func (o *option) Conn() Conn           { return o.conn }
func (o *option) EnabledForThem() bool { return telnetQYes == o.them }
func (o *option) EnabledForUs() bool   { return telnetQYes == o.us }

func (o *option) Subnegotiation(bytes []byte) {
	o.conn.Logf(DEBUG, "RECV: IAC SB %s %q IAC SE", optionByte(o.Byte()), bytes)
}

func (o *option) Update(byte, bool, bool, bool, bool) {}

func (o *option) disableThem() error {
	return o.disable(&o.them, DONT)
}

func (o *option) disableUs() error {
	return o.disable(&o.us, WONT)
}

func (o *option) disable(state *telnetQState, cmd byte) error {
	switch *state {
	case telnetQNo:
		// ignore
	case telnetQYes:
		*state = telnetQWantNoEmpty
		return o.sendOptionCommand(cmd, o.code)
	case telnetQWantNoEmpty:
		// ignore
	case telnetQWantNoOpposite:
		*state = telnetQWantNoEmpty
	case telnetQWantYesEmpty:
		*state = telnetQWantYesOpposite
	case telnetQWantYesOpposite:
		// ignore
	}
	return nil
}

func (o *option) enableThem() error {
	return o.enable(&o.them, DO)
}

func (o *option) enableUs() error {
	return o.enable(&o.us, WILL)
}
func (o *option) enable(state *telnetQState, cmd byte) error {
	switch *state {
	case telnetQNo:
		*state = telnetQWantYesEmpty
		return o.sendOptionCommand(cmd, o.code)
	case telnetQYes:
		// ignore
	case telnetQWantNoEmpty:
		*state = telnetQWantNoOpposite
	case telnetQWantNoOpposite:
		// ignore
	case telnetQWantYesEmpty:
		// ignore
	case telnetQWantYesOpposite:
		*state = telnetQWantYesEmpty
	}
	return nil
}

func (o *option) receive(c byte) error {
	switch c {
	case DO:
		return o.receiveEnableRequest(&o.us, o.allowUs, WILL, WONT)
	case DONT:
		return o.receiveDisableDemand(&o.us, WILL, WONT)
	case WILL:
		return o.receiveEnableRequest(&o.them, o.allowThem, DO, DONT)
	case WONT:
		return o.receiveDisableDemand(&o.them, DO, DONT)
	}
	return nil
}

func (o *option) receiveEnableRequest(state *telnetQState, allowed bool, accept, reject byte) error {
	switch *state {
	case telnetQNo:
		if allowed {
			*state = telnetQYes
			return o.sendOptionCommand(accept, o.code)
		} else {
			return o.sendOptionCommand(reject, o.code)
		}
	case telnetQYes:
		// ignore
	case telnetQWantNoEmpty:
		*state = telnetQNo
	case telnetQWantNoOpposite:
		*state = telnetQYes
	case telnetQWantYesEmpty:
		*state = telnetQYes
	case telnetQWantYesOpposite:
		*state = telnetQWantNoEmpty
		return o.sendOptionCommand(reject, o.code)
	}
	return nil
}

func (o *option) receiveDisableDemand(state *telnetQState, accept, reject byte) error {
	switch *state {
	case telnetQNo:
		// ignore
	case telnetQYes:
		*state = telnetQNo
		return o.sendOptionCommand(reject, o.code)
	case telnetQWantNoEmpty:
		*state = telnetQNo
	case telnetQWantNoOpposite:
		*state = telnetQWantYesEmpty
		return o.sendOptionCommand(accept, o.code)
	case telnetQWantYesEmpty:
		*state = telnetQNo
	case telnetQWantYesOpposite:
		*state = telnetQNo
	}
	return nil
}

func (o *option) sendOptionCommand(cmd, opt byte) error {
	o.Conn().Logf(DEBUG, "SEND: IAC %s %s", commandByte(cmd), optionByte(opt))
	_, err := o.Conn().Send([]byte{IAC, cmd, opt})
	return err
}

type telnetQState int

const (
	telnetQNo telnetQState = 0 + iota
	telnetQYes
	telnetQWantNoEmpty
	telnetQWantNoOpposite
	telnetQWantYesEmpty
	telnetQWantYesOpposite
)

func (s telnetQState) String() string {
	switch s {
	case telnetQNo:
		return "No"
	case telnetQYes:
		return "Yes"
	case telnetQWantNoEmpty:
		return "WantNo:Empty"
	case telnetQWantNoOpposite:
		return "WantNo:Opposite"
	case telnetQWantYesEmpty:
		return "WantYes:Empty"
	case telnetQWantYesOpposite:
		return "WantYes:Opposite"
	default:
		panic("unknown state")
	}
}
