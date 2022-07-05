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
	code               byte
	allowUs, allowThem bool
	us, them           telnetQState
}

func newOption(c byte) *option {
	return &option{code: c}
}

type sendfunc func(p []byte) error

func (o *option) disableThem(send sendfunc) error {
	return o.disable(&o.them, DONT, send)
}

func (o *option) disableUs(send sendfunc) error {
	return o.disable(&o.us, WONT, send)
}

func (o *option) disable(state *telnetQState, cmd byte, send sendfunc) error {
	switch *state {
	case telnetQNo:
		// ignore
	case telnetQYes:
		*state = telnetQWantNoEmpty
		return send([]byte{IAC, cmd, o.code})
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

func (o *option) enabledForThem() bool {
	return telnetQYes == o.them
}

func (o *option) enabledForUs() bool {
	return telnetQYes == o.us
}

func (o *option) enableThem(send sendfunc) error {
	return o.enable(&o.them, DO, send)
}

func (o *option) enableUs(send sendfunc) error {
	return o.enable(&o.us, WILL, send)
}

func (o *option) enable(state *telnetQState, cmd byte, send sendfunc) error {
	switch *state {
	case telnetQNo:
		*state = telnetQWantYesEmpty
		return send([]byte{IAC, cmd, o.code})
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

func (o *option) receive(c byte, send sendfunc) error {
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
			return send([]byte{IAC, accept, o.code})
		} else {
			return send([]byte{IAC, reject, o.code})
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
		return send([]byte{IAC, reject, o.code})
	}
	return nil
}

func (o *option) receiveDisableDemand(state *telnetQState, accept, reject byte, send sendfunc) error {
	switch *state {
	case telnetQNo:
		// ignore
	case telnetQYes:
		*state = telnetQNo
		return send([]byte{IAC, reject, o.code})
	case telnetQWantNoEmpty:
		*state = telnetQNo
	case telnetQWantNoOpposite:
		*state = telnetQWantYesEmpty
		return send([]byte{IAC, accept, o.code})
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
