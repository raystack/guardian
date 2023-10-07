// Code generated by mockery v2.33.3. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/raystack/guardian/domain"
	mock "github.com/stretchr/testify/mock"
)

// ActivityManager is an autogenerated mock type for the activityManager type
type ActivityManager struct {
	mock.Mock
}

type ActivityManager_Expecter struct {
	mock *mock.Mock
}

func (_m *ActivityManager) EXPECT() *ActivityManager_Expecter {
	return &ActivityManager_Expecter{mock: &_m.Mock}
}

// GetActivities provides a mock function with given fields: _a0, _a1, _a2
func (_m *ActivityManager) GetActivities(_a0 context.Context, _a1 domain.Provider, _a2 domain.ListActivitiesFilter) ([]*domain.Activity, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 []*domain.Activity
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, domain.Provider, domain.ListActivitiesFilter) ([]*domain.Activity, error)); ok {
		return rf(_a0, _a1, _a2)
	}
	if rf, ok := ret.Get(0).(func(context.Context, domain.Provider, domain.ListActivitiesFilter) []*domain.Activity); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Activity)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, domain.Provider, domain.ListActivitiesFilter) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ActivityManager_GetActivities_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetActivities'
type ActivityManager_GetActivities_Call struct {
	*mock.Call
}

// GetActivities is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 domain.Provider
//   - _a2 domain.ListActivitiesFilter
func (_e *ActivityManager_Expecter) GetActivities(_a0 interface{}, _a1 interface{}, _a2 interface{}) *ActivityManager_GetActivities_Call {
	return &ActivityManager_GetActivities_Call{Call: _e.mock.On("GetActivities", _a0, _a1, _a2)}
}

func (_c *ActivityManager_GetActivities_Call) Run(run func(_a0 context.Context, _a1 domain.Provider, _a2 domain.ListActivitiesFilter)) *ActivityManager_GetActivities_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.Provider), args[2].(domain.ListActivitiesFilter))
	})
	return _c
}

func (_c *ActivityManager_GetActivities_Call) Return(_a0 []*domain.Activity, _a1 error) *ActivityManager_GetActivities_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ActivityManager_GetActivities_Call) RunAndReturn(run func(context.Context, domain.Provider, domain.ListActivitiesFilter) ([]*domain.Activity, error)) *ActivityManager_GetActivities_Call {
	_c.Call.Return(run)
	return _c
}

// NewActivityManager creates a new instance of ActivityManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewActivityManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *ActivityManager {
	mock := &ActivityManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
