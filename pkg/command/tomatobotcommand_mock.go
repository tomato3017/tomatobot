// Code generated by mockery v2.43.2. DO NOT EDIT.

package command

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockTomatobotCommand is an autogenerated mock type for the TomatobotCommand type
type MockTomatobotCommand struct {
	mock.Mock
}

type MockTomatobotCommand_Expecter struct {
	mock *mock.Mock
}

func (_m *MockTomatobotCommand) EXPECT() *MockTomatobotCommand_Expecter {
	return &MockTomatobotCommand_Expecter{mock: &_m.Mock}
}

// Description provides a mock function with given fields:
func (_m *MockTomatobotCommand) Description() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Description")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockTomatobotCommand_Description_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Description'
type MockTomatobotCommand_Description_Call struct {
	*mock.Call
}

// Description is a helper method to define mock.On call
func (_e *MockTomatobotCommand_Expecter) Description() *MockTomatobotCommand_Description_Call {
	return &MockTomatobotCommand_Description_Call{Call: _e.mock.On("Description")}
}

func (_c *MockTomatobotCommand_Description_Call) Run(run func()) *MockTomatobotCommand_Description_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTomatobotCommand_Description_Call) Return(_a0 string) *MockTomatobotCommand_Description_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTomatobotCommand_Description_Call) RunAndReturn(run func() string) *MockTomatobotCommand_Description_Call {
	_c.Call.Return(run)
	return _c
}

// Execute provides a mock function with given fields: ctx, params
func (_m *MockTomatobotCommand) Execute(ctx context.Context, params CommandParams) error {
	ret := _m.Called(ctx, params)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, CommandParams) error); ok {
		r0 = rf(ctx, params)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockTomatobotCommand_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type MockTomatobotCommand_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - ctx context.Context
//   - params CommandParams
func (_e *MockTomatobotCommand_Expecter) Execute(ctx interface{}, params interface{}) *MockTomatobotCommand_Execute_Call {
	return &MockTomatobotCommand_Execute_Call{Call: _e.mock.On("Execute", ctx, params)}
}

func (_c *MockTomatobotCommand_Execute_Call) Run(run func(ctx context.Context, params CommandParams)) *MockTomatobotCommand_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(CommandParams))
	})
	return _c
}

func (_c *MockTomatobotCommand_Execute_Call) Return(_a0 error) *MockTomatobotCommand_Execute_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTomatobotCommand_Execute_Call) RunAndReturn(run func(context.Context, CommandParams) error) *MockTomatobotCommand_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// Help provides a mock function with given fields:
func (_m *MockTomatobotCommand) Help() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Help")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockTomatobotCommand_Help_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Help'
type MockTomatobotCommand_Help_Call struct {
	*mock.Call
}

// Help is a helper method to define mock.On call
func (_e *MockTomatobotCommand_Expecter) Help() *MockTomatobotCommand_Help_Call {
	return &MockTomatobotCommand_Help_Call{Call: _e.mock.On("Help")}
}

func (_c *MockTomatobotCommand_Help_Call) Run(run func()) *MockTomatobotCommand_Help_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTomatobotCommand_Help_Call) Return(_a0 string) *MockTomatobotCommand_Help_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTomatobotCommand_Help_Call) RunAndReturn(run func() string) *MockTomatobotCommand_Help_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockTomatobotCommand creates a new instance of MockTomatobotCommand. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockTomatobotCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTomatobotCommand {
	mock := &MockTomatobotCommand{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}