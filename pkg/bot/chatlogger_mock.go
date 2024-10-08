// Code generated by mockery v2.43.2. DO NOT EDIT.

package bot

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	tgapi "github.com/tomato3017/tomatobot/pkg/bot/models/tgapi"
)

// MockChatLogger is an autogenerated mock type for the ChatLogger type
type MockChatLogger struct {
	mock.Mock
}

type MockChatLogger_Expecter struct {
	mock *mock.Mock
}

func (_m *MockChatLogger) EXPECT() *MockChatLogger_Expecter {
	return &MockChatLogger_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields:
func (_m *MockChatLogger) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockChatLogger_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type MockChatLogger_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *MockChatLogger_Expecter) Close() *MockChatLogger_Close_Call {
	return &MockChatLogger_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *MockChatLogger_Close_Call) Run(run func()) *MockChatLogger_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockChatLogger_Close_Call) Return(_a0 error) *MockChatLogger_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockChatLogger_Close_Call) RunAndReturn(run func() error) *MockChatLogger_Close_Call {
	_c.Call.Return(run)
	return _c
}

// LogChats provides a mock function with given fields: ctx, msg
func (_m *MockChatLogger) LogChats(ctx context.Context, msg tgapi.TGBotMsg) error {
	ret := _m.Called(ctx, msg)

	if len(ret) == 0 {
		panic("no return value specified for LogChats")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, tgapi.TGBotMsg) error); ok {
		r0 = rf(ctx, msg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockChatLogger_LogChats_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LogChats'
type MockChatLogger_LogChats_Call struct {
	*mock.Call
}

// LogChats is a helper method to define mock.On call
//   - ctx context.Context
//   - msg tgapi.TGBotMsg
func (_e *MockChatLogger_Expecter) LogChats(ctx interface{}, msg interface{}) *MockChatLogger_LogChats_Call {
	return &MockChatLogger_LogChats_Call{Call: _e.mock.On("LogChats", ctx, msg)}
}

func (_c *MockChatLogger_LogChats_Call) Run(run func(ctx context.Context, msg tgapi.TGBotMsg)) *MockChatLogger_LogChats_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(tgapi.TGBotMsg))
	})
	return _c
}

func (_c *MockChatLogger_LogChats_Call) Return(_a0 error) *MockChatLogger_LogChats_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockChatLogger_LogChats_Call) RunAndReturn(run func(context.Context, tgapi.TGBotMsg) error) *MockChatLogger_LogChats_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockChatLogger creates a new instance of MockChatLogger. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockChatLogger(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockChatLogger {
	mock := &MockChatLogger{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
