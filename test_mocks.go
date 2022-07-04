// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stesla/telnet (interfaces: Conn,OptionHandler)

// Package telnet is a generated GoMock package.
package telnet

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	encoding "golang.org/x/text/encoding"
)

// MockConn is a mock of Conn interface.
type MockConn struct {
	ctrl     *gomock.Controller
	recorder *MockConnMockRecorder
}

// MockConnMockRecorder is the mock recorder for MockConn.
type MockConnMockRecorder struct {
	mock *MockConn
}

// NewMockConn creates a new mock instance.
func NewMockConn(ctrl *gomock.Controller) *MockConn {
	mock := &MockConn{ctrl: ctrl}
	mock.recorder = &MockConnMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConn) EXPECT() *MockConnMockRecorder {
	return m.recorder
}

// AllowOption mocks base method.
func (m *MockConn) AllowOption(arg0 OptionHandler, arg1, arg2 bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AllowOption", arg0, arg1, arg2)
}

// AllowOption indicates an expected call of AllowOption.
func (mr *MockConnMockRecorder) AllowOption(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllowOption", reflect.TypeOf((*MockConn)(nil).AllowOption), arg0, arg1, arg2)
}

// EnableOptionForThem mocks base method.
func (m *MockConn) EnableOptionForThem(arg0 byte, arg1 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnableOptionForThem", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// EnableOptionForThem indicates an expected call of EnableOptionForThem.
func (mr *MockConnMockRecorder) EnableOptionForThem(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnableOptionForThem", reflect.TypeOf((*MockConn)(nil).EnableOptionForThem), arg0, arg1)
}

// EnableOptionForUs mocks base method.
func (m *MockConn) EnableOptionForUs(arg0 byte, arg1 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnableOptionForUs", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// EnableOptionForUs indicates an expected call of EnableOptionForUs.
func (mr *MockConnMockRecorder) EnableOptionForUs(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnableOptionForUs", reflect.TypeOf((*MockConn)(nil).EnableOptionForUs), arg0, arg1)
}

// Read mocks base method.
func (m *MockConn) Read(arg0 []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockConnMockRecorder) Read(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockConn)(nil).Read), arg0)
}

// Send mocks base method.
func (m *MockConn) Send(arg0 ...byte) (int, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Send", varargs...)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Send indicates an expected call of Send.
func (mr *MockConnMockRecorder) Send(arg0 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockConn)(nil).Send), arg0...)
}

// SetEncoding mocks base method.
func (m *MockConn) SetEncoding(arg0 encoding.Encoding) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetEncoding", arg0)
}

// SetEncoding indicates an expected call of SetEncoding.
func (mr *MockConnMockRecorder) SetEncoding(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetEncoding", reflect.TypeOf((*MockConn)(nil).SetEncoding), arg0)
}

// SuppressGoAhead mocks base method.
func (m *MockConn) SuppressGoAhead(arg0 bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SuppressGoAhead", arg0)
}

// SuppressGoAhead indicates an expected call of SuppressGoAhead.
func (mr *MockConnMockRecorder) SuppressGoAhead(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SuppressGoAhead", reflect.TypeOf((*MockConn)(nil).SuppressGoAhead), arg0)
}

// Write mocks base method.
func (m *MockConn) Write(arg0 []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockConnMockRecorder) Write(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockConn)(nil).Write), arg0)
}

// MockOptionHandler is a mock of OptionHandler interface.
type MockOptionHandler struct {
	ctrl     *gomock.Controller
	recorder *MockOptionHandlerMockRecorder
}

// MockOptionHandlerMockRecorder is the mock recorder for MockOptionHandler.
type MockOptionHandlerMockRecorder struct {
	mock *MockOptionHandler
}

// NewMockOptionHandler creates a new mock instance.
func NewMockOptionHandler(ctrl *gomock.Controller) *MockOptionHandler {
	mock := &MockOptionHandler{ctrl: ctrl}
	mock.recorder = &MockOptionHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOptionHandler) EXPECT() *MockOptionHandlerMockRecorder {
	return m.recorder
}

// Disable mocks base method.
func (m *MockOptionHandler) Disable(arg0 Conn) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Disable", arg0)
}

// Disable indicates an expected call of Disable.
func (mr *MockOptionHandlerMockRecorder) Disable(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Disable", reflect.TypeOf((*MockOptionHandler)(nil).Disable), arg0)
}

// Enable mocks base method.
func (m *MockOptionHandler) Enable(arg0 Conn) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Enable", arg0)
}

// Enable indicates an expected call of Enable.
func (mr *MockOptionHandlerMockRecorder) Enable(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enable", reflect.TypeOf((*MockOptionHandler)(nil).Enable), arg0)
}

// Option mocks base method.
func (m *MockOptionHandler) Option() byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Option")
	ret0, _ := ret[0].(byte)
	return ret0
}

// Option indicates an expected call of Option.
func (mr *MockOptionHandlerMockRecorder) Option() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Option", reflect.TypeOf((*MockOptionHandler)(nil).Option))
}

// Subnegotiation mocks base method.
func (m *MockOptionHandler) Subnegotiation(arg0 Conn, arg1 []byte) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Subnegotiation", arg0, arg1)
}

// Subnegotiation indicates an expected call of Subnegotiation.
func (mr *MockOptionHandlerMockRecorder) Subnegotiation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subnegotiation", reflect.TypeOf((*MockOptionHandler)(nil).Subnegotiation), arg0, arg1)
}