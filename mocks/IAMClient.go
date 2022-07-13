// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	"github.com/odpf/guardian/domain"
	mock "github.com/stretchr/testify/mock"
)

// IAMClient is an autogenerated mock type for the IAMClient type
type IAMClient struct {
	mock.Mock
}

// GetUser provides a mock function with given fields: id
func (_m *IAMClient) GetUser(id string) (interface{}, error) {
	ret := _m.Called(id)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(string) interface{}); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsActiveUser provides a mock function with given fields: id
func (_m *IAMClient) IsActiveUser(id string, config *domain.IAMConfig) (bool, error) {
	ret := _m.Called(id, config)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, *domain.IAMConfig) bool); ok {
		r0 = rf(id, config)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *domain.IAMConfig) error); ok {
		r1 = rf(id, config)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
