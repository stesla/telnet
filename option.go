package telnet

type Option interface {
	Allow(them, us bool)
	Byte() byte
	EnabledForThem() bool
	EnabledForUs() bool
	Receive(c byte, send sendfunc) error
	Subnegotiation(Conn, []byte)
	Update(c Conn, option byte, theyChanged, them, weChanged, us bool)

	disableThem(transmitter) error
	disableUs(transmitter) error
	enableThem(transmitter) error
	enableUs(transmitter) error
}

func newOptionMap() *optionMap {
	return &optionMap{
		m: make(map[byte]Option),
	}
}

type optionMap struct {
	m map[byte]Option
}

func (m *optionMap) each(fn func(Option)) {
	for _, opt := range m.m {
		fn(opt)
	}
}

func (m *optionMap) get(c byte) (opt Option) {
	opt, ok := m.m[c]
	if !ok {
		opt = NewOption(c)
		m.m[c] = opt
	}
	return
}

func (m *optionMap) put(o Option) {
	m.m[o.Byte()] = o
}

type option struct {
	code               byte
	allowUs, allowThem bool
	us, them           telnetQState
}

func NewOption(c byte) *option {
	return &option{code: c}
}

func (o *option) Allow(them, us bool)  { o.allowThem, o.allowUs = them, us }
func (o *option) Byte() byte           { return o.code }
func (o *option) EnabledForThem() bool { return telnetQYes == o.them }
func (o *option) EnabledForUs() bool   { return telnetQYes == o.us }

func (o *option) Subnegotiation(c Conn, bytes []byte) {
	c.Logf(DEBUG, "RECV: IAC SB %s %q IAC SE", optionByte(o.Byte()), bytes)
}
func (o *option) Update(Conn, byte, bool, bool, bool, bool) {}

type sendfunc func(cmd, opt byte) error

type transmitter interface {
	sendOptionCommand(cmd, opt byte) error
}

func (o *option) disableThem(tx transmitter) error {
	return o.disable(&o.them, DONT, tx.sendOptionCommand)
}

func (o *option) disableUs(tx transmitter) error {
	return o.disable(&o.us, WONT, tx.sendOptionCommand)
}

func (o *option) disable(state *telnetQState, cmd byte, send sendfunc) error {
	switch *state {
	case telnetQNo:
		// ignore
	case telnetQYes:
		*state = telnetQWantNoEmpty
		return send(cmd, o.code)
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

func (o *option) enableThem(tx transmitter) error {
	return o.enable(&o.them, DO, tx.sendOptionCommand)
}

func (o *option) enableUs(tx transmitter) error {
	return o.enable(&o.us, WILL, tx.sendOptionCommand)
}

func (o *option) enable(state *telnetQState, cmd byte, send sendfunc) error {
	switch *state {
	case telnetQNo:
		*state = telnetQWantYesEmpty
		return send(cmd, o.code)
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

func (o *option) Receive(c byte, send sendfunc) error {
	switch c {
	case DO:
		return o.receiveEnableRequest(&o.us, o.allowUs, WILL, WONT, send)
	case DONT:
		return o.receiveDisableDemand(&o.us, WILL, WONT, send)
	case WILL:
		return o.receiveEnableRequest(&o.them, o.allowThem, DO, DONT, send)
	case WONT:
		return o.receiveDisableDemand(&o.them, DO, DONT, send)
	}
	return nil
}

func (o *option) receiveEnableRequest(state *telnetQState, allowed bool, accept, reject byte, send sendfunc) error {
	switch *state {
	case telnetQNo:
		if allowed {
			*state = telnetQYes
			return send(accept, o.code)
		} else {
			return send(reject, o.code)
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
		return send(reject, o.code)
	}
	return nil
}

func (o *option) receiveDisableDemand(state *telnetQState, accept, reject byte, send sendfunc) error {
	switch *state {
	case telnetQNo:
		// ignore
	case telnetQYes:
		*state = telnetQNo
		return send(reject, o.code)
	case telnetQWantNoEmpty:
		*state = telnetQNo
	case telnetQWantNoOpposite:
		*state = telnetQWantYesEmpty
		return send(accept, o.code)
	case telnetQWantYesEmpty:
		*state = telnetQNo
	case telnetQWantYesOpposite:
		*state = telnetQNo
	}
	return nil
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
