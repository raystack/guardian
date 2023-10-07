// Code generated by mockery v2.33.3. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/raystack/guardian/domain"
	mock "github.com/stretchr/testify/mock"
)

// ResourceService is an autogenerated mock type for the resourceService type
type ResourceService struct {
	mock.Mock
}

type ResourceService_Expecter struct {
	mock *mock.Mock
}

func (_m *ResourceService) EXPECT() *ResourceService_Expecter {
	return &ResourceService_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: _a0, _a1
func (_m *ResourceService) Get(_a0 context.Context, _a1 *domain.ResourceIdentifier) (*domain.Resource, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *domain.Resource
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.ResourceIdentifier) (*domain.Resource, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *domain.ResourceIdentifier) *domain.Resource); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Resource)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *domain.ResourceIdentifier) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ResourceService_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type ResourceService_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *domain.ResourceIdentifier
func (_e *ResourceService_Expecter) Get(_a0 interface{}, _a1 interface{}) *ResourceService_Get_Call {
	return &ResourceService_Get_Call{Call: _e.mock.On("Get", _a0, _a1)}
}

func (_c *ResourceService_Get_Call) Run(run func(_a0 context.Context, _a1 *domain.ResourceIdentifier)) *ResourceService_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*domain.ResourceIdentifier))
	})
	return _c
}

func (_c *ResourceService_Get_Call) Return(_a0 *domain.Resource, _a1 error) *ResourceService_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ResourceService_Get_Call) RunAndReturn(run func(context.Context, *domain.ResourceIdentifier) (*domain.Resource, error)) *ResourceService_Get_Call {
	_c.Call.Return(run)
	return _c
}

// NewResourceService creates a new instance of ResourceService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewResourceService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ResourceService {
	mock := &ResourceService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
