// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	domain "github.com/odpf/guardian/domain"
	mock "github.com/stretchr/testify/mock"
)

// AppealService is an autogenerated mock type for the AppealService type
type AppealService struct {
	mock.Mock
}

// Cancel provides a mock function with given fields: _a0
func (_m *AppealService) Cancel(_a0 uint) (*domain.Appeal, error) {
	ret := _m.Called(_a0)

	var r0 *domain.Appeal
	if rf, ok := ret.Get(0).(func(uint) *domain.Appeal); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Appeal)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Create provides a mock function with given fields: _a0
func (_m *AppealService) Create(_a0 []*domain.Appeal) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func([]*domain.Appeal) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Find provides a mock function with given fields: _a0
func (_m *AppealService) Find(_a0 map[string]interface{}) ([]*domain.Appeal, error) {
	ret := _m.Called(_a0)

	var r0 []*domain.Appeal
	if rf, ok := ret.Get(0).(func(map[string]interface{}) []*domain.Appeal); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Appeal)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(map[string]interface{}) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: _a0
func (_m *AppealService) GetByID(_a0 uint) (*domain.Appeal, error) {
	ret := _m.Called(_a0)

	var r0 *domain.Appeal
	if rf, ok := ret.Get(0).(func(uint) *domain.Appeal); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Appeal)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MakeAction provides a mock function with given fields: _a0
func (_m *AppealService) MakeAction(_a0 domain.ApprovalAction) (*domain.Appeal, error) {
	ret := _m.Called(_a0)

	var r0 *domain.Appeal
	if rf, ok := ret.Get(0).(func(domain.ApprovalAction) *domain.Appeal); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Appeal)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(domain.ApprovalAction) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Revoke provides a mock function with given fields: id, actor
func (_m *AppealService) Revoke(id uint, actor string) (*domain.Appeal, error) {
	ret := _m.Called(id, actor)

	var r0 *domain.Appeal
	if rf, ok := ret.Get(0).(func(uint, string) *domain.Appeal); ok {
		r0 = rf(id, actor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Appeal)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint, string) error); ok {
		r1 = rf(id, actor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
