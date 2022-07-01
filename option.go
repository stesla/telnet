package telnet

func newOptionMap() *optionMap {
	return &optionMap{
		m: make(map[byte]*option),
	}
}

type optionMap struct {
	m map[byte]*option
}

func (m *optionMap) get(c byte) (opt *option) {
	opt, ok := m.m[c]
	if !ok {
		opt = newOption(c)
		m.m[c] = opt
	}
	return
}

type option struct {
	code byte

	allowUs, allowThem bool
	us, them           telnetQState
}

func newOption(c byte) *option {
	return &option{code: c}
}

func (o *option) receive(c byte, fn func(byte)) {
	switch c {
	case DO:
		o.receiveEnableRequest(&o.us, o.allowUs, WILL, WONT, fn)
	case DONT:
		o.receiveDisableDemand(&o.us, WILL, WONT, fn)
	case WILL:
		o.receiveEnableRequest(&o.them, o.allowThem, DO, DONT, fn)
	case WONT:
		o.receiveDisableDemand(&o.them, DO, DONT, fn)
	}
}

func (o *option) receiveEnableRequest(state *telnetQState, allowed bool, accept, reject byte, fn func(byte)) {
	switch *state {
	case telnetQNo:
		if allowed {
			*state = telnetQYes
			fn(accept)
		} else {
			fn(reject)
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
		fn(reject)
	}
}

func (o *option) receiveDisableDemand(state *telnetQState, accept, reject byte, fn func(byte)) {
	switch *state {
	case telnetQNo:
		// ignore
	case telnetQYes:
		*state = telnetQNo
		fn(reject)
	case telnetQWantNoEmpty:
		*state = telnetQNo
	case telnetQWantNoOpposite:
		*state = telnetQWantYesEmpty
		fn(accept)
	case telnetQWantYesEmpty:
		*state = telnetQNo
	case telnetQWantYesOpposite:
		*state = telnetQNo
	}
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
