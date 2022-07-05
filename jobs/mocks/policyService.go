// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/odpf/guardian/domain"

	mock "github.com/stretchr/testify/mock"
)

// PolicyService is an autogenerated mock type for the policyService type
type PolicyService struct {
	mock.Mock
}

// GetOne provides a mock function with given fields: ctx, id, version
func (_m *PolicyService) GetOne(ctx context.Context, id string, version uint) (*domain.Policy, error) {
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
