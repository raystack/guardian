// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/raystack/guardian/domain"
	mock "github.com/stretchr/testify/mock"
)

// Repository is an autogenerated mock type for the repository type
type Repository struct {
	mock.Mock
}

type Repository_Expecter struct {
	mock *mock.Mock
}

func (_m *Repository) EXPECT() *Repository_Expecter {
	return &Repository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: _a0, _a1
func (_m *Repository) Create(_a0 context.Context, _a1 *domain.Policy) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Policy) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Repository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type Repository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//  - _a0 context.Context
//  - _a1 *domain.Policy
func (_e *Repository_Expecter) Create(_a0 interface{}, _a1 interface{}) *Repository_Create_Call {
	return &Repository_Create_Call{Call: _e.mock.On("Create", _a0, _a1)}
}

func (_c *Repository_Create_Call) Run(run func(_a0 context.Context, _a1 *domain.Policy)) *Repository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*domain.Policy))
	})
	return _c
}

func (_c *Repository_Create_Call) Return(_a0 error) *Repository_Create_Call {
	_c.Call.Return(_a0)
	return _c
}

// Find provides a mock function with given fields: _a0
func (_m *Repository) Find(_a0 context.Context) ([]*domain.Policy, error) {
	ret := _m.Called(_a0)

	var r0 []*domain.Policy
	if rf, ok := ret.Get(0).(func(context.Context) []*domain.Policy); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Policy)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Repository_Find_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Find'
type Repository_Find_Call struct {
	*mock.Call
}

// Find is a helper method to define mock.On call
//  - _a0 context.Context
func (_e *Repository_Expecter) Find(_a0 interface{}) *Repository_Find_Call {
	return &Repository_Find_Call{Call: _e.mock.On("Find", _a0)}
}

func (_c *Repository_Find_Call) Run(run func(_a0 context.Context)) *Repository_Find_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Repository_Find_Call) Return(_a0 []*domain.Policy, _a1 error) *Repository_Find_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetOne provides a mock function with given fields: ctx, id, version
func (_m *Repository) GetOne(ctx context.Context, id string, version uint) (*domain.Policy, error) {
	ret := _m.Called(ctx, id, version)

	var r0 *domain.Policy
	if rf, ok := ret.Get(0).(func(context.Context, string, uint) *domain.Policy); ok {
		r0 = rf(ctx, id, version)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Policy)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, uint) error); ok {
		r1 = rf(ctx, id, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Repository_GetOne_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOne'
type Repository_GetOne_Call struct {
	*mock.Call
}

// GetOne is a helper method to define mock.On call
//  - ctx context.Context
//  - id string
//  - version uint
func (_e *Repository_Expecter) GetOne(ctx interface{}, id interface{}, version interface{}) *Repository_GetOne_Call {
	return &Repository_GetOne_Call{Call: _e.mock.On("GetOne", ctx, id, version)}
}

func (_c *Repository_GetOne_Call) Run(run func(ctx context.Context, id string, version uint)) *Repository_GetOne_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(uint))
	})
	return _c
}

func (_c *Repository_GetOne_Call) Return(_a0 *domain.Policy, _a1 error) *Repository_GetOne_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}
