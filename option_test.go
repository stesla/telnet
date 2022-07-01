package telnet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type qMethodTest struct {
	start, end telnetQState
	permitted  bool
	expected   byte
}

func TestQMethodReceiveDO(t *testing.T) {
	tests := []*qMethodTest{
		&qMethodTest{start: telnetQNo, permitted: false, end: telnetQNo, expected: WONT},
		&qMethodTest{start: telnetQNo, permitted: true, end: telnetQYes, expected: WILL},
		&qMethodTest{start: telnetQYes, end: telnetQYes},
		&qMethodTest{start: telnetQWantNoEmpty, end: telnetQNo},
		&qMethodTest{start: telnetQWantNoOpposite, end: telnetQYes},
		&qMethodTest{start: telnetQWantYesEmpty, end: telnetQYes},
		&qMethodTest{start: telnetQWantYesOpposite, end: telnetQWantNoEmpty, expected: WONT},
	}
	for i, q := range tests {
		o := newOption(SuppressGoAhead)
		o.us, o.allowUs = q.start, q.permitted
		var called bool
		o.receive(DO, func(c byte) {
			called = true
			assert.Equal(t, q.expected, c, "test %d", i)
		})
		assert.Equal(t, q.expected != 0, called, "test %d", i)
		assert.Equal(t, q.end, o.us, "test %d", i)
	}
}

func TestQMethodReceiveDONT(t *testing.T) {
	tests := []*qMethodTest{
		&qMethodTest{start: telnetQNo, end: telnetQNo},
		&qMethodTest{start: telnetQYes, end: telnetQNo, expected: WONT},
		&qMethodTest{start: telnetQWantNoEmpty, end: telnetQNo},
		&qMethodTest{start: telnetQWantNoOpposite, end: telnetQWantYesEmpty, expected: WILL},
		&qMethodTest{start: telnetQWantYesEmpty, end: telnetQNo},
		&qMethodTest{start: telnetQWantYesOpposite, end: telnetQNo},
	}
	for i, q := range tests {
		o := newOption(SuppressGoAhead)
		o.us, o.allowUs = q.start, q.permitted
		var called bool
		o.receive(DONT, func(c byte) {
			called = true
			assert.Equal(t, q.expected, c, "test %d", i)
		})
		assert.Equal(t, q.expected != 0, called, "test %d", i)
		assert.Equal(t, q.end, o.us, "test %d", i)
	}
}

func TestQMethodReceiveWILL(t *testing.T) {
	tests := []*qMethodTest{
		&qMethodTest{start: telnetQNo, permitted: false, end: telnetQNo, expected: DONT},
		&qMethodTest{start: telnetQNo, permitted: true, end: telnetQYes, expected: DO},
		&qMethodTest{start: telnetQYes, end: telnetQYes},
		&qMethodTest{start: telnetQWantNoEmpty, end: telnetQNo},
		&qMethodTest{start: telnetQWantNoOpposite, end: telnetQYes},
		&qMethodTest{start: telnetQWantYesEmpty, end: telnetQYes},
		&qMethodTest{start: telnetQWantYesOpposite, end: telnetQWantNoEmpty, expected: DONT},
	}
	for i, q := range tests {
		o := newOption(SuppressGoAhead)
		o.them, o.allowThem = q.start, q.permitted
		var called bool
		o.receive(WILL, func(c byte) {
			called = true
			assert.Equal(t, q.expected, c, "test %d", i)
		})
		assert.Equal(t, q.expected != 0, called, "test %d", i)
		assert.Equal(t, q.end, o.them, "test %d", i)
	}
}

func TestQMethodReceiveWONT(t *testing.T) {
	tests := []*qMethodTest{
		//;&qMethodTest{start: telnetQNo, end: telnetQNo},
		&qMethodTest{start: telnetQYes, end: telnetQNo, expected: DONT},
		//&qMethodTest{start: telnetQWantNoEmpty, end: telnetQNo},
		&qMethodTest{start: telnetQWantNoOpposite, end: telnetQWantYesEmpty, expected: DO},
		//&qMethodTest{start: telnetQWantYesEmpty, end: telnetQNo},
		//&qMethodTest{start: telnetQWantYesOpposite, end: telnetQNo},
	}
	for i, q := range tests {
		o := newOption(SuppressGoAhead)
		o.them, o.allowThem = q.start, q.permitted
		var called bool
		o.receive(WONT, func(c byte) {
			called = true
			assert.Equal(t, q.expected, c, "test %d", i)
		})
		assert.Equal(t, q.expected != 0, called, "test %d", i)
		assert.Equal(t, q.end, o.them, "test %d", i)

	}
}
