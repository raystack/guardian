// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/odpf/guardian/domain"
	mock "github.com/stretchr/testify/mock"
)

// ResourceService is an autogenerated mock type for the resourceService type
type ResourceService struct {
	mock.Mock
}

// BulkUpsert provides a mock function with given fields: _a0, _a1
func (_m *ResourceService) BulkUpsert(_a0 context.Context, _a1 []*domain.Resource) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []*domain.Resource) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Find provides a mock function with given fields: _a0, _a1
func (_m *ResourceService) Find(_a0 context.Context, _a1 map[string]interface{}) ([]*domain.Resource, error) {
	ret := _m.Called(_a0, _a1)

	var r0 []*domain.Resource
	if rf, ok := ret.Get(0).(func(context.Context, map[string]interface{}) []*domain.Resource); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Resource)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, map[string]interface{}) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
