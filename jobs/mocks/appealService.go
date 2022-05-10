// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	domain "github.com/odpf/guardian/domain"

	mock "github.com/stretchr/testify/mock"
)

// AppealService is an autogenerated mock type for the appealService type
type AppealService struct {
	mock.Mock
}

// Find provides a mock function with given fields: _a0
func (_m *AppealService) Find(_a0 *domain.ListAppealsFilter) ([]*domain.Appeal, error) {
	ret := _m.Called(_a0)

	var r0 []*domain.Appeal
	if rf, ok := ret.Get(0).(func(*domain.ListAppealsFilter) []*domain.Appeal); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Appeal)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*domain.ListAppealsFilter) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Revoke provides a mock function with given fields: id, actor, reason
func (_m *AppealService) Revoke(id string, actor string, reason string) (*domain.Appeal, error) {
	ret := _m.Called(id, actor, reason)

	var r0 *domain.Appeal
	if rf, ok := ret.Get(0).(func(string, string, string) *domain.Appeal); ok {
		r0 = rf(id, actor, reason)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Appeal)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(id, actor, reason)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
