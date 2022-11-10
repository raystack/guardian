// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	domain "github.com/odpf/guardian/domain"
	mock "github.com/stretchr/testify/mock"
)

// Notifier is an autogenerated mock type for the notifier type
type Notifier struct {
	mock.Mock
}

type Notifier_Expecter struct {
	mock *mock.Mock
}

func (_m *Notifier) EXPECT() *Notifier_Expecter {
	return &Notifier_Expecter{mock: &_m.Mock}
}

// Notify provides a mock function with given fields: _a0
func (_m *Notifier) Notify(_a0 []domain.Notification) []error {
	ret := _m.Called(_a0)

	var r0 []error
	if rf, ok := ret.Get(0).(func([]domain.Notification) []error); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]error)
		}
	}

	return r0
}

// Notifier_Notify_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Notify'
type Notifier_Notify_Call struct {
	*mock.Call
}

// Notify is a helper method to define mock.On call
//  - _a0 []domain.Notification
func (_e *Notifier_Expecter) Notify(_a0 interface{}) *Notifier_Notify_Call {
	return &Notifier_Notify_Call{Call: _e.mock.On("Notify", _a0)}
}

func (_c *Notifier_Notify_Call) Run(run func(_a0 []domain.Notification)) *Notifier_Notify_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]domain.Notification))
	})
	return _c
}

func (_c *Notifier_Notify_Call) Return(_a0 []error) *Notifier_Notify_Call {
	_c.Call.Return(_a0)
	return _c
}

type mockConstructorTestingTNewNotifier interface {
	mock.TestingT
	Cleanup(func())
}

// NewNotifier creates a new instance of Notifier. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewNotifier(t mockConstructorTestingTNewNotifier) *Notifier {
	mock := &Notifier{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
