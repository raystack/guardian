// Code generated by mockery v2.33.3. DO NOT EDIT.

package mocks

import (
	context "context"

	grant "github.com/raystack/guardian/core/grant"
	domain "github.com/raystack/guardian/domain"

	mock "github.com/stretchr/testify/mock"
)

// GrantService is an autogenerated mock type for the grantService type
type GrantService struct {
	mock.Mock
}

type GrantService_Expecter struct {
	mock *mock.Mock
}

func (_m *GrantService) EXPECT() *GrantService_Expecter {
	return &GrantService_Expecter{mock: &_m.Mock}
}

// BulkRevoke provides a mock function with given fields: ctx, filter, actor, reason
func (_m *GrantService) BulkRevoke(ctx context.Context, filter domain.RevokeGrantsFilter, actor string, reason string) ([]*domain.Grant, error) {
	ret := _m.Called(ctx, filter, actor, reason)

	var r0 []*domain.Grant
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, domain.RevokeGrantsFilter, string, string) ([]*domain.Grant, error)); ok {
		return rf(ctx, filter, actor, reason)
	}
	if rf, ok := ret.Get(0).(func(context.Context, domain.RevokeGrantsFilter, string, string) []*domain.Grant); ok {
		r0 = rf(ctx, filter, actor, reason)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Grant)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, domain.RevokeGrantsFilter, string, string) error); ok {
		r1 = rf(ctx, filter, actor, reason)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GrantService_BulkRevoke_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BulkRevoke'
type GrantService_BulkRevoke_Call struct {
	*mock.Call
}

// BulkRevoke is a helper method to define mock.On call
//   - ctx context.Context
//   - filter domain.RevokeGrantsFilter
//   - actor string
//   - reason string
func (_e *GrantService_Expecter) BulkRevoke(ctx interface{}, filter interface{}, actor interface{}, reason interface{}) *GrantService_BulkRevoke_Call {
	return &GrantService_BulkRevoke_Call{Call: _e.mock.On("BulkRevoke", ctx, filter, actor, reason)}
}

func (_c *GrantService_BulkRevoke_Call) Run(run func(ctx context.Context, filter domain.RevokeGrantsFilter, actor string, reason string)) *GrantService_BulkRevoke_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.RevokeGrantsFilter), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *GrantService_BulkRevoke_Call) Return(_a0 []*domain.Grant, _a1 error) *GrantService_BulkRevoke_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *GrantService_BulkRevoke_Call) RunAndReturn(run func(context.Context, domain.RevokeGrantsFilter, string, string) ([]*domain.Grant, error)) *GrantService_BulkRevoke_Call {
	_c.Call.Return(run)
	return _c
}

// DormancyCheck provides a mock function with given fields: _a0, _a1
func (_m *GrantService) DormancyCheck(_a0 context.Context, _a1 domain.DormancyCheckCriteria) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, domain.DormancyCheckCriteria) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GrantService_DormancyCheck_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DormancyCheck'
type GrantService_DormancyCheck_Call struct {
	*mock.Call
}

// DormancyCheck is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 domain.DormancyCheckCriteria
func (_e *GrantService_Expecter) DormancyCheck(_a0 interface{}, _a1 interface{}) *GrantService_DormancyCheck_Call {
	return &GrantService_DormancyCheck_Call{Call: _e.mock.On("DormancyCheck", _a0, _a1)}
}

func (_c *GrantService_DormancyCheck_Call) Run(run func(_a0 context.Context, _a1 domain.DormancyCheckCriteria)) *GrantService_DormancyCheck_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.DormancyCheckCriteria))
	})
	return _c
}

func (_c *GrantService_DormancyCheck_Call) Return(_a0 error) *GrantService_DormancyCheck_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GrantService_DormancyCheck_Call) RunAndReturn(run func(context.Context, domain.DormancyCheckCriteria) error) *GrantService_DormancyCheck_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with given fields: _a0, _a1
func (_m *GrantService) List(_a0 context.Context, _a1 domain.ListGrantsFilter) ([]domain.Grant, error) {
	ret := _m.Called(_a0, _a1)

	var r0 []domain.Grant
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, domain.ListGrantsFilter) ([]domain.Grant, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, domain.ListGrantsFilter) []domain.Grant); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]domain.Grant)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, domain.ListGrantsFilter) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GrantService_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type GrantService_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 domain.ListGrantsFilter
func (_e *GrantService_Expecter) List(_a0 interface{}, _a1 interface{}) *GrantService_List_Call {
	return &GrantService_List_Call{Call: _e.mock.On("List", _a0, _a1)}
}

func (_c *GrantService_List_Call) Run(run func(_a0 context.Context, _a1 domain.ListGrantsFilter)) *GrantService_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.ListGrantsFilter))
	})
	return _c
}

func (_c *GrantService_List_Call) Return(_a0 []domain.Grant, _a1 error) *GrantService_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *GrantService_List_Call) RunAndReturn(run func(context.Context, domain.ListGrantsFilter) ([]domain.Grant, error)) *GrantService_List_Call {
	_c.Call.Return(run)
	return _c
}

// Revoke provides a mock function with given fields: ctx, id, actor, reason, opts
func (_m *GrantService) Revoke(ctx context.Context, id string, actor string, reason string, opts ...grant.Option) (*domain.Grant, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, id, actor, reason)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *domain.Grant
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, ...grant.Option) (*domain.Grant, error)); ok {
		return rf(ctx, id, actor, reason, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, ...grant.Option) *domain.Grant); ok {
		r0 = rf(ctx, id, actor, reason, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Grant)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, ...grant.Option) error); ok {
		r1 = rf(ctx, id, actor, reason, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GrantService_Revoke_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Revoke'
type GrantService_Revoke_Call struct {
	*mock.Call
}

// Revoke is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
//   - actor string
//   - reason string
//   - opts ...grant.Option
func (_e *GrantService_Expecter) Revoke(ctx interface{}, id interface{}, actor interface{}, reason interface{}, opts ...interface{}) *GrantService_Revoke_Call {
	return &GrantService_Revoke_Call{Call: _e.mock.On("Revoke",
		append([]interface{}{ctx, id, actor, reason}, opts...)...)}
}

func (_c *GrantService_Revoke_Call) Run(run func(ctx context.Context, id string, actor string, reason string, opts ...grant.Option)) *GrantService_Revoke_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grant.Option, len(args)-4)
		for i, a := range args[4:] {
			if a != nil {
				variadicArgs[i] = a.(grant.Option)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(string), variadicArgs...)
	})
	return _c
}

func (_c *GrantService_Revoke_Call) Return(_a0 *domain.Grant, _a1 error) *GrantService_Revoke_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *GrantService_Revoke_Call) RunAndReturn(run func(context.Context, string, string, string, ...grant.Option) (*domain.Grant, error)) *GrantService_Revoke_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: _a0, _a1
func (_m *GrantService) Update(_a0 context.Context, _a1 *domain.Grant) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Grant) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GrantService_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type GrantService_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *domain.Grant
func (_e *GrantService_Expecter) Update(_a0 interface{}, _a1 interface{}) *GrantService_Update_Call {
	return &GrantService_Update_Call{Call: _e.mock.On("Update", _a0, _a1)}
}

func (_c *GrantService_Update_Call) Run(run func(_a0 context.Context, _a1 *domain.Grant)) *GrantService_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*domain.Grant))
	})
	return _c
}

func (_c *GrantService_Update_Call) Return(_a0 error) *GrantService_Update_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GrantService_Update_Call) RunAndReturn(run func(context.Context, *domain.Grant) error) *GrantService_Update_Call {
	_c.Call.Return(run)
	return _c
}

// NewGrantService creates a new instance of GrantService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGrantService(t interface {
	mock.TestingT
	Cleanup(func())
}) *GrantService {
	mock := &GrantService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
