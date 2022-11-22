// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/odpf/guardian/domain"
	mock "github.com/stretchr/testify/mock"
)

// ProviderActivityService is an autogenerated mock type for the providerActivityService type
type ProviderActivityService struct {
	mock.Mock
}

type ProviderActivityService_Expecter struct {
	mock *mock.Mock
}

func (_m *ProviderActivityService) EXPECT() *ProviderActivityService_Expecter {
	return &ProviderActivityService_Expecter{mock: &_m.Mock}
}

// BulkInsert provides a mock function with given fields: _a0, _a1
func (_m *ProviderActivityService) BulkInsert(_a0 context.Context, _a1 []*domain.ProviderActivity) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []*domain.ProviderActivity) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ProviderActivityService_BulkInsert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BulkInsert'
type ProviderActivityService_BulkInsert_Call struct {
	*mock.Call
}

// BulkInsert is a helper method to define mock.On call
//  - _a0 context.Context
//  - _a1 []*domain.ProviderActivity
func (_e *ProviderActivityService_Expecter) BulkInsert(_a0 interface{}, _a1 interface{}) *ProviderActivityService_BulkInsert_Call {
	return &ProviderActivityService_BulkInsert_Call{Call: _e.mock.On("BulkInsert", _a0, _a1)}
}

func (_c *ProviderActivityService_BulkInsert_Call) Run(run func(_a0 context.Context, _a1 []*domain.ProviderActivity)) *ProviderActivityService_BulkInsert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]*domain.ProviderActivity))
	})
	return _c
}

func (_c *ProviderActivityService_BulkInsert_Call) Return(_a0 error) *ProviderActivityService_BulkInsert_Call {
	_c.Call.Return(_a0)
	return _c
}
