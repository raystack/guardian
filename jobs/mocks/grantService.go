// Code generated by mockery v2.32.4. DO NOT EDIT.

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
