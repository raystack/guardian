// Code generated by mockery v2.38.0. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/raystack/guardian/domain"
	mock "github.com/stretchr/testify/mock"
)

// PolicyService is an autogenerated mock type for the policyService type
type PolicyService struct {
	mock.Mock
}

type PolicyService_Expecter struct {
	mock *mock.Mock
}

func (_m *PolicyService) EXPECT() *PolicyService_Expecter {
	return &PolicyService_Expecter{mock: &_m.Mock}
}

// Find provides a mock function with given fields: _a0
func (_m *PolicyService) Find(_a0 context.Context) ([]*domain.Policy, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Find")
	}

	var r0 []*domain.Policy
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*domain.Policy, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*domain.Policy); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Policy)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PolicyService_Find_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Find'
type PolicyService_Find_Call struct {
	*mock.Call
}

// Find is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *PolicyService_Expecter) Find(_a0 interface{}) *PolicyService_Find_Call {
	return &PolicyService_Find_Call{Call: _e.mock.On("Find", _a0)}
}

func (_c *PolicyService_Find_Call) Run(run func(_a0 context.Context)) *PolicyService_Find_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *PolicyService_Find_Call) Return(_a0 []*domain.Policy, _a1 error) *PolicyService_Find_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *PolicyService_Find_Call) RunAndReturn(run func(context.Context) ([]*domain.Policy, error)) *PolicyService_Find_Call {
	_c.Call.Return(run)
	return _c
}

// GetOne provides a mock function with given fields: _a0, _a1, _a2
func (_m *PolicyService) GetOne(_a0 context.Context, _a1 string, _a2 uint) (*domain.Policy, error) {
	ret := _m.Called(_a0, _a1, _a2)

	if len(ret) == 0 {
		panic("no return value specified for GetOne")
	}

	var r0 *domain.Policy
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, uint) (*domain.Policy, error)); ok {
		return rf(_a0, _a1, _a2)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, uint) *domain.Policy); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Policy)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, uint) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PolicyService_GetOne_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOne'
type PolicyService_GetOne_Call struct {
	*mock.Call
}

// GetOne is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 string
//   - _a2 uint
func (_e *PolicyService_Expecter) GetOne(_a0 interface{}, _a1 interface{}, _a2 interface{}) *PolicyService_GetOne_Call {
	return &PolicyService_GetOne_Call{Call: _e.mock.On("GetOne", _a0, _a1, _a2)}
}

func (_c *PolicyService_GetOne_Call) Run(run func(_a0 context.Context, _a1 string, _a2 uint)) *PolicyService_GetOne_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(uint))
	})
	return _c
}

func (_c *PolicyService_GetOne_Call) Return(_a0 *domain.Policy, _a1 error) *PolicyService_GetOne_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *PolicyService_GetOne_Call) RunAndReturn(run func(context.Context, string, uint) (*domain.Policy, error)) *PolicyService_GetOne_Call {
	_c.Call.Return(run)
	return _c
}

// NewPolicyService creates a new instance of PolicyService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPolicyService(t interface {
	mock.TestingT
	Cleanup(func())
}) *PolicyService {
	mock := &PolicyService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
