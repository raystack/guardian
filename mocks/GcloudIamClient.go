// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	gcloudiam "github.com/odpf/guardian/provider/gcloudiam"
	mock "github.com/stretchr/testify/mock"
)

// GcloudIamClient is an autogenerated mock type for the GcloudIamClient type
type GcloudIamClient struct {
	mock.Mock
}

// GetRoles provides a mock function with given fields:
func (_m *GcloudIamClient) GetRoles() ([]*gcloudiam.Role, error) {
	ret := _m.Called()

	var r0 []*gcloudiam.Role
	if rf, ok := ret.Get(0).(func() []*gcloudiam.Role); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*gcloudiam.Role)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GrantAccess provides a mock function with given fields: accountType, accountID, role
func (_m *GcloudIamClient) GrantAccess(accountType string, accountID string, role string) error {
	ret := _m.Called(accountType, accountID, role)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(accountType, accountID, role)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RevokeAccess provides a mock function with given fields: accountType, accountID, role
func (_m *GcloudIamClient) RevokeAccess(accountType string, accountID string, role string) error {
	ret := _m.Called(accountType, accountID, role)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(accountType, accountID, role)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
