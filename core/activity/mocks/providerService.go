// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/raystack/guardian/domain"
	mock "github.com/stretchr/testify/mock"
)

// ProviderService is an autogenerated mock type for the providerService type
type ProviderService struct {
	mock.Mock
}

type ProviderService_Expecter struct {
	mock *mock.Mock
}

func (_m *ProviderService) EXPECT() *ProviderService_Expecter {
	return &ProviderService_Expecter{mock: &_m.Mock}
}

// ImportActivities provides a mock function with given fields: _a0, _a1
func (_m *ProviderService) ImportActivities(_a0 context.Context, _a1 domain.ImportActivitiesFilter) ([]*domain.Activity, error) {
	ret := _m.Called(_a0, _a1)

	var r0 []*domain.Activity
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, domain.ImportActivitiesFilter) ([]*domain.Activity, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, domain.ImportActivitiesFilter) []*domain.Activity); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Activity)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, domain.ImportActivitiesFilter) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProviderService_ImportActivities_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ImportActivities'
type ProviderService_ImportActivities_Call struct {
	*mock.Call
}

// ImportActivities is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 domain.ImportActivitiesFilter
func (_e *ProviderService_Expecter) ImportActivities(_a0 interface{}, _a1 interface{}) *ProviderService_ImportActivities_Call {
	return &ProviderService_ImportActivities_Call{Call: _e.mock.On("ImportActivities", _a0, _a1)}
}

func (_c *ProviderService_ImportActivities_Call) Run(run func(_a0 context.Context, _a1 domain.ImportActivitiesFilter)) *ProviderService_ImportActivities_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.ImportActivitiesFilter))
	})
	return _c
}

func (_c *ProviderService_ImportActivities_Call) Return(_a0 []*domain.Activity, _a1 error) *ProviderService_ImportActivities_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProviderService_ImportActivities_Call) RunAndReturn(run func(context.Context, domain.ImportActivitiesFilter) ([]*domain.Activity, error)) *ProviderService_ImportActivities_Call {
	_c.Call.Return(run)
	return _c
}

// NewProviderService creates a new instance of ProviderService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProviderService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ProviderService {
	mock := &ProviderService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
