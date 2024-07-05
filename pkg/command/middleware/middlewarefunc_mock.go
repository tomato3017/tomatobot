// Code generated by mockery v2.43.2. DO NOT EDIT.

package middleware

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	models "github.com/tomato3017/tomatobot/pkg/command/models"
)

// MockMiddlewareFunc is an autogenerated mock type for the MiddlewareFunc type
type MockMiddlewareFunc struct {
	mock.Mock
}

type MockMiddlewareFunc_Expecter struct {
	mock *mock.Mock
}

func (_m *MockMiddlewareFunc) EXPECT() *MockMiddlewareFunc_Expecter {
	return &MockMiddlewareFunc_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: ctx, params
func (_m *MockMiddlewareFunc) Execute(ctx context.Context, params models.CommandParams) error {
	ret := _m.Called(ctx, params)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, models.CommandParams) error); ok {
		r0 = rf(ctx, params)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockMiddlewareFunc_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type MockMiddlewareFunc_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - ctx context.Context
//   - params models.CommandParams
func (_e *MockMiddlewareFunc_Expecter) Execute(ctx interface{}, params interface{}) *MockMiddlewareFunc_Execute_Call {
	return &MockMiddlewareFunc_Execute_Call{Call: _e.mock.On("Execute", ctx, params)}
}

func (_c *MockMiddlewareFunc_Execute_Call) Run(run func(ctx context.Context, params models.CommandParams)) *MockMiddlewareFunc_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(models.CommandParams))
	})
	return _c
}

func (_c *MockMiddlewareFunc_Execute_Call) Return(_a0 error) *MockMiddlewareFunc_Execute_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockMiddlewareFunc_Execute_Call) RunAndReturn(run func(context.Context, models.CommandParams) error) *MockMiddlewareFunc_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockMiddlewareFunc creates a new instance of MockMiddlewareFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockMiddlewareFunc(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMiddlewareFunc {
	mock := &MockMiddlewareFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
