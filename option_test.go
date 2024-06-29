package telnet

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	fn func() error
}

func TestQMethodReceive(t *testing.T) {
	conn := NewMockConn(t)
	sink := NewMockEventSink(t)

	o := NewOption(SuppressGoAhead)
	o.Bind(conn, sink)

	tests := []*qMethodTest{
		{state: &o.us, allow: &o.allowUs, receive: DO, start: telnetQNo, permitted: false, end: telnetQNo, expected: WONT},
		{state: &o.us, allow: &o.allowUs, receive: DO, start: telnetQNo, permitted: true, end: telnetQYes, expected: WILL},
		{state: &o.us, allow: &o.allowUs, receive: DO, start: telnetQYes, end: telnetQYes},
		{state: &o.us, allow: &o.allowUs, receive: DO, start: telnetQWantNoEmpty, end: telnetQNo},
		{state: &o.us, allow: &o.allowUs, receive: DO, start: telnetQWantNoOpposite, end: telnetQYes},
		{state: &o.us, allow: &o.allowUs, receive: DO, start: telnetQWantYesEmpty, end: telnetQYes},
		{state: &o.us, allow: &o.allowUs, receive: DO, start: telnetQWantYesOpposite, end: telnetQWantNoEmpty, expected: WONT},
		{state: &o.us, allow: &o.allowUs, receive: DONT, start: telnetQNo, end: telnetQNo},
		{state: &o.us, allow: &o.allowUs, receive: DONT, start: telnetQYes, end: telnetQNo, expected: WONT},
		{state: &o.us, allow: &o.allowUs, receive: DONT, start: telnetQWantNoEmpty, end: telnetQNo},
		{state: &o.us, allow: &o.allowUs, receive: DONT, start: telnetQWantNoOpposite, end: telnetQWantYesEmpty, expected: WILL},
		{state: &o.us, allow: &o.allowUs, receive: DONT, start: telnetQWantYesEmpty, end: telnetQNo},
		{state: &o.us, allow: &o.allowUs, receive: DONT, start: telnetQWantYesOpposite, end: telnetQNo},
		{state: &o.them, allow: &o.allowThem, receive: WILL, start: telnetQNo, permitted: false, end: telnetQNo, expected: DONT},
		{state: &o.them, allow: &o.allowThem, receive: WILL, start: telnetQNo, permitted: true, end: telnetQYes, expected: DO},
		{state: &o.them, allow: &o.allowThem, receive: WILL, start: telnetQYes, end: telnetQYes},
		{state: &o.them, allow: &o.allowThem, receive: WILL, start: telnetQWantNoEmpty, end: telnetQNo},
		{state: &o.them, allow: &o.allowThem, receive: WILL, start: telnetQWantNoOpposite, end: telnetQYes},
		{state: &o.them, allow: &o.allowThem, receive: WILL, start: telnetQWantYesEmpty, end: telnetQYes},
		{state: &o.them, allow: &o.allowThem, receive: WILL, start: telnetQWantYesOpposite, end: telnetQWantNoEmpty, expected: DONT},
		{state: &o.them, allow: &o.allowThem, receive: WONT, start: telnetQNo, end: telnetQNo},
		{state: &o.them, allow: &o.allowThem, receive: WONT, start: telnetQYes, end: telnetQNo, expected: DONT},
		{state: &o.them, allow: &o.allowThem, receive: WONT, start: telnetQWantNoEmpty, end: telnetQNo},
		{state: &o.them, allow: &o.allowThem, receive: WONT, start: telnetQWantNoOpposite, end: telnetQWantYesEmpty, expected: DO},
		{state: &o.them, allow: &o.allowThem, receive: WONT, start: telnetQWantYesEmpty, end: telnetQNo},
		{state: &o.them, allow: &o.allowThem, receive: WONT, start: telnetQWantYesOpposite, end: telnetQNo},
	}
	for _, q := range tests {
		o.us, o.them = telnetQNo, telnetQNo
		*q.state, *q.allow = q.start, q.permitted
		if q.expected != 0 {
			conn.EXPECT().Logf(mock.Anything, commandByte(q.expected), optionByte(o.code))
			expected := []byte{IAC, q.expected, o.code}
			conn.EXPECT().Send(expected).Return(len(expected), nil)
		}
		if (q.start != telnetQYes && q.end == telnetQYes) || (q.start == telnetQYes && q.end != telnetQYes) {
			if q.state == &o.them {
				sink.EXPECT().SendEvent("update-option", UpdateOptionEvent{o, true, false})
			} else if q.state == &o.us {
				sink.EXPECT().SendEvent("update-option", UpdateOptionEvent{o, false, true})
			}
		}
		o.receive(q.receive)
		testMsg := fmt.Sprintf("test %s %s %v", commandByte(q.receive), q.start, q.permitted)
		assert.Equal(t, q.end, *q.state, testMsg)
	}
}

func TestQMethodEnableOrDisable(t *testing.T) {
	conn := NewMockConn(t)

	o := NewOption(SuppressGoAhead)
	o.Bind(conn, nil)

	disableThem := fmt.Sprintf("%p", o.disableThem)
	disableUs := fmt.Sprintf("%p", o.disableUs)
	enableThem := fmt.Sprintf("%p", o.enableThem)
	enableUs := fmt.Sprintf("%p", o.enableUs)
	tests := []*qMethodTest{
		{fn: o.disableThem, state: &o.them, start: telnetQNo, end: telnetQNo},
		{fn: o.disableThem, state: &o.them, start: telnetQYes, end: telnetQWantNoEmpty, expected: DONT},
		{fn: o.disableThem, state: &o.them, start: telnetQWantNoEmpty, end: telnetQWantNoEmpty},
		{fn: o.disableThem, state: &o.them, start: telnetQWantNoOpposite, end: telnetQWantNoEmpty},
		{fn: o.disableThem, state: &o.them, start: telnetQWantYesEmpty, end: telnetQWantYesOpposite},
		{fn: o.disableThem, state: &o.them, start: telnetQWantYesOpposite, end: telnetQWantYesOpposite},
		{fn: o.disableUs, state: &o.us, start: telnetQNo, end: telnetQNo},
		{fn: o.disableUs, state: &o.us, start: telnetQYes, end: telnetQWantNoEmpty, expected: WONT},
		{fn: o.disableUs, state: &o.us, start: telnetQWantNoEmpty, end: telnetQWantNoEmpty},
		{fn: o.disableUs, state: &o.us, start: telnetQWantNoOpposite, end: telnetQWantNoEmpty},
		{fn: o.disableUs, state: &o.us, start: telnetQWantYesEmpty, end: telnetQWantYesOpposite},
		{fn: o.disableUs, state: &o.us, start: telnetQWantYesOpposite, end: telnetQWantYesOpposite},
		{fn: o.enableThem, state: &o.them, start: telnetQNo, end: telnetQWantYesEmpty, expected: DO},
		{fn: o.enableThem, state: &o.them, start: telnetQYes, end: telnetQYes},
		{fn: o.enableThem, state: &o.them, start: telnetQWantNoEmpty, end: telnetQWantNoOpposite},
		{fn: o.enableThem, state: &o.them, start: telnetQWantNoOpposite, end: telnetQWantNoOpposite},
		{fn: o.enableThem, state: &o.them, start: telnetQWantYesEmpty, end: telnetQWantYesEmpty},
		{fn: o.enableThem, state: &o.them, start: telnetQWantYesOpposite, end: telnetQWantYesEmpty},
		{fn: o.enableUs, state: &o.us, start: telnetQNo, end: telnetQWantYesEmpty, expected: WILL},
		{fn: o.enableUs, state: &o.us, start: telnetQYes, end: telnetQYes},
		{fn: o.enableUs, state: &o.us, start: telnetQWantNoEmpty, end: telnetQWantNoOpposite},
		{fn: o.enableUs, state: &o.us, start: telnetQWantNoOpposite, end: telnetQWantNoOpposite},
		{fn: o.enableUs, state: &o.us, start: telnetQWantYesEmpty, end: telnetQWantYesEmpty},
		{fn: o.enableUs, state: &o.us, start: telnetQWantYesOpposite, end: telnetQWantYesEmpty},
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
		if q.expected != 0 {
			conn.EXPECT().Logf(mock.Anything, commandByte(q.expected), optionByte(o.code))
			expected := []byte{IAC, q.expected, o.code}
			conn.EXPECT().Send(expected).Return(len(expected), nil)
		}
		err := q.fn()
		assert.NoError(t, err, testMsg)
	}
}
