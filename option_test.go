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
		&qMethodTest{start: telnetQNo, end: telnetQNo},
		&qMethodTest{start: telnetQYes, end: telnetQNo, expected: DONT},
		&qMethodTest{start: telnetQWantNoEmpty, end: telnetQNo},
		&qMethodTest{start: telnetQWantNoOpposite, end: telnetQWantYesEmpty, expected: DO},
		&qMethodTest{start: telnetQWantYesEmpty, end: telnetQNo},
		&qMethodTest{start: telnetQWantYesOpposite, end: telnetQNo},
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

func TestQMethodAskEnableThem(t *testing.T) {
	tests := []*qMethodTest{
		&qMethodTest{start: telnetQNo, end: telnetQWantYesEmpty, expected: DO},
		&qMethodTest{start: telnetQYes, end: telnetQYes},
		&qMethodTest{start: telnetQWantNoEmpty, end: telnetQWantNoOpposite},
		&qMethodTest{start: telnetQWantNoOpposite, end: telnetQWantNoOpposite},
		&qMethodTest{start: telnetQWantYesEmpty, end: telnetQWantYesEmpty},
		&qMethodTest{start: telnetQWantYesOpposite, end: telnetQWantYesEmpty},
	}
	for _, q := range tests {
		o := newOption(SuppressGoAhead)
		o.them = q.start
		called := false
		err := o.enableThem(func(p ...byte) error {
			called = true
			if q.expected != 0 {
				assert.Equal(t, []byte{IAC, q.expected, SuppressGoAhead}, p)
			}
			return nil
		})
		if q.expected != 0 {
			assert.True(t, called)
		}
		assert.NoError(t, err)
	}
}

func TestQMethodDisableThem(t *testing.T) {
	tests := []*qMethodTest{
		&qMethodTest{start: telnetQNo, end: telnetQNo},
		&qMethodTest{start: telnetQYes, end: telnetQWantNoEmpty, expected: DONT},
		&qMethodTest{start: telnetQWantNoEmpty, end: telnetQWantNoEmpty},
		&qMethodTest{start: telnetQWantNoOpposite, end: telnetQWantNoEmpty},
		&qMethodTest{start: telnetQWantYesEmpty, end: telnetQWantYesOpposite},
		&qMethodTest{start: telnetQWantYesOpposite, end: telnetQWantYesOpposite},
	}
	for _, q := range tests {
		o := newOption(SuppressGoAhead)
		o.them = q.start
		called := false
		err := o.disableThem(func(p ...byte) error {
			called = true
			if q.expected != 0 {
				assert.Equal(t, []byte{IAC, q.expected, SuppressGoAhead}, p)
			}
			return nil
		})
		if q.expected != 0 {
			assert.True(t, called)
		}
		assert.NoError(t, err)
	}
}

func TestQMethodEnableUs(t *testing.T) {
	tests := []*qMethodTest{
		&qMethodTest{start: telnetQNo, end: telnetQWantYesEmpty, expected: WILL},
		&qMethodTest{start: telnetQYes, end: telnetQYes},
		&qMethodTest{start: telnetQWantNoEmpty, end: telnetQWantNoOpposite},
		&qMethodTest{start: telnetQWantNoOpposite, end: telnetQWantNoOpposite},
		&qMethodTest{start: telnetQWantYesEmpty, end: telnetQWantYesEmpty},
		&qMethodTest{start: telnetQWantYesOpposite, end: telnetQWantYesEmpty},
	}
	for _, q := range tests {
		o := newOption(SuppressGoAhead)
		o.us = q.start
		called := false
		err := o.enableUs(func(p ...byte) error {
			called = true
			if q.expected != 0 {
				assert.Equal(t, []byte{IAC, q.expected, SuppressGoAhead}, p)
			}
			return nil
		})
		if q.expected != 0 {
			assert.True(t, called)
		}
		assert.NoError(t, err)
	}
}

func TestQMethodDisableUs(t *testing.T) {
	tests := []*qMethodTest{
		&qMethodTest{start: telnetQNo, end: telnetQNo},
		&qMethodTest{start: telnetQYes, end: telnetQWantNoEmpty, expected: WONT},
		&qMethodTest{start: telnetQWantNoEmpty, end: telnetQWantNoEmpty},
		&qMethodTest{start: telnetQWantNoOpposite, end: telnetQWantNoEmpty},
		&qMethodTest{start: telnetQWantYesEmpty, end: telnetQWantYesOpposite},
		&qMethodTest{start: telnetQWantYesOpposite, end: telnetQWantYesOpposite},
	}
	for _, q := range tests {
		o := newOption(SuppressGoAhead)
		o.us = q.start
		called := false
		err := o.disableUs(func(p ...byte) error {
			called = true
			if q.expected != 0 {
				assert.Equal(t, []byte{IAC, q.expected, SuppressGoAhead}, p)
			}
			return nil
		})
		if q.expected != 0 {
			assert.True(t, called)
		}
		assert.NoError(t, err)
	}
}
