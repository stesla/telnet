// Code generated by mockery v2.43.2. DO NOT EDIT.

package telnet

import mock "github.com/stretchr/testify/mock"

// MockEventListener is an autogenerated mock type for the EventListener type
type MockEventListener struct {
	mock.Mock
}

type MockEventListener_Expecter struct {
	mock *mock.Mock
}

func (_m *MockEventListener) EXPECT() *MockEventListener_Expecter {
	return &MockEventListener_Expecter{mock: &_m.Mock}
}

// HandleEvent provides a mock function with given fields: _a0
func (_m *MockEventListener) HandleEvent(_a0 interface{}) {
	_m.Called(_a0)
}

// MockEventListener_HandleEvent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HandleEvent'
type MockEventListener_HandleEvent_Call struct {
	*mock.Call
}

// HandleEvent is a helper method to define mock.On call
//   - _a0 interface{}
func (_e *MockEventListener_Expecter) HandleEvent(_a0 interface{}) *MockEventListener_HandleEvent_Call {
	return &MockEventListener_HandleEvent_Call{Call: _e.mock.On("HandleEvent", _a0)}
}

func (_c *MockEventListener_HandleEvent_Call) Run(run func(_a0 interface{})) *MockEventListener_HandleEvent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *MockEventListener_HandleEvent_Call) Return() *MockEventListener_HandleEvent_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockEventListener_HandleEvent_Call) RunAndReturn(run func(interface{})) *MockEventListener_HandleEvent_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockEventListener creates a new instance of MockEventListener. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockEventListener(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockEventListener {
	mock := &MockEventListener{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}