// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/odpf/guardian/domain"
	mock "github.com/stretchr/testify/mock"
)

// AccessService is an autogenerated mock type for the accessService type
type AccessService struct {
	mock.Mock
}

type AccessService_Expecter struct {
	mock *mock.Mock
}

func (_m *AccessService) EXPECT() *AccessService_Expecter {
	return &AccessService_Expecter{mock: &_m.Mock}
}

// GetByID provides a mock function with given fields: _a0, _a1
func (_m *AccessService) GetByID(_a0 context.Context, _a1 string) (*domain.Access, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *domain.Access
	if rf, ok := ret.Get(0).(func(context.Context, string) *domain.Access); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Access)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AccessService_GetByID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByID'
type AccessService_GetByID_Call struct {
	*mock.Call
}

// GetByID is a helper method to define mock.On call
//  - _a0 context.Context
//  - _a1 string
func (_e *AccessService_Expecter) GetByID(_a0 interface{}, _a1 interface{}) *AccessService_GetByID_Call {
	return &AccessService_GetByID_Call{Call: _e.mock.On("GetByID", _a0, _a1)}
}

func (_c *AccessService_GetByID_Call) Run(run func(_a0 context.Context, _a1 string)) *AccessService_GetByID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *AccessService_GetByID_Call) Return(_a0 *domain.Access, _a1 error) *AccessService_GetByID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// List provides a mock function with given fields: _a0, _a1
func (_m *AccessService) List(_a0 context.Context, _a1 domain.ListAccessesFilter) ([]domain.Access, error) {
	ret := _m.Called(_a0, _a1)

	var r0 []domain.Access
	if rf, ok := ret.Get(0).(func(context.Context, domain.ListAccessesFilter) []domain.Access); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]domain.Access)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, domain.ListAccessesFilter) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AccessService_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type AccessService_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//  - _a0 context.Context
//  - _a1 domain.ListAccessesFilter
func (_e *AccessService_Expecter) List(_a0 interface{}, _a1 interface{}) *AccessService_List_Call {
	return &AccessService_List_Call{Call: _e.mock.On("List", _a0, _a1)}
}

func (_c *AccessService_List_Call) Run(run func(_a0 context.Context, _a1 domain.ListAccessesFilter)) *AccessService_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.ListAccessesFilter))
	})
	return _c
}

func (_c *AccessService_List_Call) Return(_a0 []domain.Access, _a1 error) *AccessService_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}