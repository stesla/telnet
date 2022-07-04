package telnet

import (
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type qMethodTest struct {
	start, end telnetQState
	permitted  bool
	expected   byte

	state *telnetQState

	// for the receive test
	receive byte
	allow   *bool

	// for the enable/disable test
	fn func(sendfunc) error
}

func themFn(opt *option, state telnetQState, allowed bool) {
	opt.them, opt.allowThem = state, allowed
}

func usFn(opt *option, state telnetQState, allowed bool) {
	opt.us, opt.allowUs = state, allowed
}

func TestQMethodReceive(t *testing.T) {
	o := newOption(SuppressGoAhead)
	tests := []*qMethodTest{
		&qMethodTest{state: &o.us, allow: &o.allowUs, receive: DO, start: telnetQNo, permitted: false, end: telnetQNo, expected: WONT},
		&qMethodTest{state: &o.us, allow: &o.allowUs, receive: DO, start: telnetQNo, permitted: true, end: telnetQYes, expected: WILL},
		&qMethodTest{state: &o.us, allow: &o.allowUs, receive: DO, start: telnetQYes, end: telnetQYes},
		&qMethodTest{state: &o.us, allow: &o.allowUs, receive: DO, start: telnetQWantNoEmpty, end: telnetQNo},
		&qMethodTest{state: &o.us, allow: &o.allowUs, receive: DO, start: telnetQWantNoOpposite, end: telnetQYes},
		&qMethodTest{state: &o.us, allow: &o.allowUs, receive: DO, start: telnetQWantYesEmpty, end: telnetQYes},
		&qMethodTest{state: &o.us, allow: &o.allowUs, receive: DO, start: telnetQWantYesOpposite, end: telnetQWantNoEmpty, expected: WONT},
		&qMethodTest{state: &o.us, allow: &o.allowUs, receive: DONT, start: telnetQNo, end: telnetQNo},
		&qMethodTest{state: &o.us, allow: &o.allowUs, receive: DONT, start: telnetQYes, end: telnetQNo, expected: WONT},
		&qMethodTest{state: &o.us, allow: &o.allowUs, receive: DONT, start: telnetQWantNoEmpty, end: telnetQNo},
		&qMethodTest{state: &o.us, allow: &o.allowUs, receive: DONT, start: telnetQWantNoOpposite, end: telnetQWantYesEmpty, expected: WILL},
		&qMethodTest{state: &o.us, allow: &o.allowUs, receive: DONT, start: telnetQWantYesEmpty, end: telnetQNo},
		&qMethodTest{state: &o.us, allow: &o.allowUs, receive: DONT, start: telnetQWantYesOpposite, end: telnetQNo},
		&qMethodTest{state: &o.them, allow: &o.allowThem, receive: WILL, start: telnetQNo, permitted: false, end: telnetQNo, expected: DONT},
		&qMethodTest{state: &o.them, allow: &o.allowThem, receive: WILL, start: telnetQNo, permitted: true, end: telnetQYes, expected: DO},
		&qMethodTest{state: &o.them, allow: &o.allowThem, receive: WILL, start: telnetQYes, end: telnetQYes},
		&qMethodTest{state: &o.them, allow: &o.allowThem, receive: WILL, start: telnetQWantNoEmpty, end: telnetQNo},
		&qMethodTest{state: &o.them, allow: &o.allowThem, receive: WILL, start: telnetQWantNoOpposite, end: telnetQYes},
		&qMethodTest{state: &o.them, allow: &o.allowThem, receive: WILL, start: telnetQWantYesEmpty, end: telnetQYes},
		&qMethodTest{state: &o.them, allow: &o.allowThem, receive: WILL, start: telnetQWantYesOpposite, end: telnetQWantNoEmpty, expected: DONT},
		&qMethodTest{state: &o.them, allow: &o.allowThem, receive: WONT, start: telnetQNo, end: telnetQNo},
		&qMethodTest{state: &o.them, allow: &o.allowThem, receive: WONT, start: telnetQYes, end: telnetQNo, expected: DONT},
		&qMethodTest{state: &o.them, allow: &o.allowThem, receive: WONT, start: telnetQWantNoEmpty, end: telnetQNo},
		&qMethodTest{state: &o.them, allow: &o.allowThem, receive: WONT, start: telnetQWantNoOpposite, end: telnetQWantYesEmpty, expected: DO},
		&qMethodTest{state: &o.them, allow: &o.allowThem, receive: WONT, start: telnetQWantYesEmpty, end: telnetQNo},
		&qMethodTest{state: &o.them, allow: &o.allowThem, receive: WONT, start: telnetQWantYesOpposite, end: telnetQNo},
	}
	for _, q := range tests {
		testMsg := fmt.Sprintf("test %s %s %v", commandByte(q.receive), q.start, q.permitted)
		*q.state, *q.allow = q.start, q.permitted
		var called bool
		o.receive(q.receive, func(c byte) {
			called = true
			assert.Equal(t, q.expected, c, testMsg)
		})
		assert.Equal(t, q.expected != 0, called, testMsg)
		assert.Equal(t, q.end, *q.state, testMsg)
	}
}

func TestQMethodEnableOrDisable(t *testing.T) {
	o := newOption(SuppressGoAhead)
	disableThem := fmt.Sprintf("%p", o.disableThem)
	disableUs := fmt.Sprintf("%p", o.disableUs)
	enableThem := fmt.Sprintf("%p", o.enableThem)
	enableUs := fmt.Sprintf("%p", o.enableUs)
	tests := []*qMethodTest{
		&qMethodTest{fn: o.disableThem, state: &o.them, start: telnetQNo, end: telnetQNo},
		&qMethodTest{fn: o.disableThem, state: &o.them, start: telnetQYes, end: telnetQWantNoEmpty, expected: DONT},
		&qMethodTest{fn: o.disableThem, state: &o.them, start: telnetQWantNoEmpty, end: telnetQWantNoEmpty},
		&qMethodTest{fn: o.disableThem, state: &o.them, start: telnetQWantNoOpposite, end: telnetQWantNoEmpty},
		&qMethodTest{fn: o.disableThem, state: &o.them, start: telnetQWantYesEmpty, end: telnetQWantYesOpposite},
		&qMethodTest{fn: o.disableThem, state: &o.them, start: telnetQWantYesOpposite, end: telnetQWantYesOpposite},
		&qMethodTest{fn: o.disableUs, state: &o.us, start: telnetQNo, end: telnetQNo},
		&qMethodTest{fn: o.disableUs, state: &o.us, start: telnetQYes, end: telnetQWantNoEmpty, expected: WONT},
		&qMethodTest{fn: o.disableUs, state: &o.us, start: telnetQWantNoEmpty, end: telnetQWantNoEmpty},
		&qMethodTest{fn: o.disableUs, state: &o.us, start: telnetQWantNoOpposite, end: telnetQWantNoEmpty},
		&qMethodTest{fn: o.disableUs, state: &o.us, start: telnetQWantYesEmpty, end: telnetQWantYesOpposite},
		&qMethodTest{fn: o.disableUs, state: &o.us, start: telnetQWantYesOpposite, end: telnetQWantYesOpposite},
		&qMethodTest{fn: o.enableThem, state: &o.them, start: telnetQNo, end: telnetQWantYesEmpty, expected: DO},
		&qMethodTest{fn: o.enableThem, state: &o.them, start: telnetQYes, end: telnetQYes},
		&qMethodTest{fn: o.enableThem, state: &o.them, start: telnetQWantNoEmpty, end: telnetQWantNoOpposite},
		&qMethodTest{fn: o.enableThem, state: &o.them, start: telnetQWantNoOpposite, end: telnetQWantNoOpposite},
		&qMethodTest{fn: o.enableThem, state: &o.them, start: telnetQWantYesEmpty, end: telnetQWantYesEmpty},
		&qMethodTest{fn: o.enableThem, state: &o.them, start: telnetQWantYesOpposite, end: telnetQWantYesEmpty},
		&qMethodTest{fn: o.enableUs, state: &o.us, start: telnetQNo, end: telnetQWantYesEmpty, expected: WILL},
		&qMethodTest{fn: o.enableUs, state: &o.us, start: telnetQYes, end: telnetQYes},
		&qMethodTest{fn: o.enableUs, state: &o.us, start: telnetQWantNoEmpty, end: telnetQWantNoOpposite},
		&qMethodTest{fn: o.enableUs, state: &o.us, start: telnetQWantNoOpposite, end: telnetQWantNoOpposite},
		&qMethodTest{fn: o.enableUs, state: &o.us, start: telnetQWantYesEmpty, end: telnetQWantYesEmpty},
		&qMethodTest{fn: o.enableUs, state: &o.us, start: telnetQWantYesOpposite, end: telnetQWantYesEmpty},
	}
	for _, q := range tests {
		var action, who string
		switch fmt.Sprintf("%p", q.fn) {
		case disableThem:
			action, who = "disable", "them"
		case disableUs:
			action, who = "disable", "us"
		case enableThem:
			action, who = "enable", "them"
		case enableUs:
			action, who = "enable", "us"
		}
		testMsg := fmt.Sprintf("test %s %s %s", action, who, q.start)
		*q.state = q.start
		called := false
		err := q.fn(func(p ...byte) error {
			called = true
			if q.expected != 0 {
				assert.Equal(t, []byte{IAC, q.expected, SuppressGoAhead}, p, testMsg)
			}
			return nil
		})
		if q.expected != 0 {
			assert.True(t, called, testMsg)
		}
		assert.NoError(t, err, testMsg)
	}
}

func TestSuppresGoAhead(t *testing.T) {
	var h OptionHandler = SuppressGoAheadOption{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	conn := NewMockConn(ctrl)

	conn.EXPECT().SuppressGoAhead(true)
	h.Enable(conn)

	conn.EXPECT().SuppressGoAhead(false)
	h.Disable(conn)

	// this should do nothing
	h.Subnegotiation(conn, []byte("foo"))
}
