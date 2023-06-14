// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	context "context"

	bigquery "github.com/raystack/guardian/plugins/providers/bigquery"

	mock "github.com/stretchr/testify/mock"
)

// CloudLoggingClientI is an autogenerated mock type for the cloudLoggingClientI type
type CloudLoggingClientI struct {
	mock.Mock
}

type CloudLoggingClientI_Expecter struct {
	mock *mock.Mock
}

func (_m *CloudLoggingClientI) EXPECT() *CloudLoggingClientI_Expecter {
	return &CloudLoggingClientI_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields:
func (_m *CloudLoggingClientI) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CloudLoggingClientI_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type CloudLoggingClientI_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *CloudLoggingClientI_Expecter) Close() *CloudLoggingClientI_Close_Call {
	return &CloudLoggingClientI_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *CloudLoggingClientI_Close_Call) Run(run func()) *CloudLoggingClientI_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *CloudLoggingClientI_Close_Call) Return(_a0 error) *CloudLoggingClientI_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

// ListLogEntries provides a mock function with given fields: _a0, _a1
func (_m *CloudLoggingClientI) ListLogEntries(_a0 context.Context, _a1 bigquery.ImportActivitiesFilter) ([]*bigquery.Activity, error) {
	ret := _m.Called(_a0, _a1)

	var r0 []*bigquery.Activity
	if rf, ok := ret.Get(0).(func(context.Context, bigquery.ImportActivitiesFilter) []*bigquery.Activity); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*bigquery.Activity)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, bigquery.ImportActivitiesFilter) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CloudLoggingClientI_ListLogEntries_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListLogEntries'
type CloudLoggingClientI_ListLogEntries_Call struct {
	*mock.Call
}

// ListLogEntries is a helper method to define mock.On call
//  - _a0 context.Context
//  - _a1 bigquery.ImportActivitiesFilter
func (_e *CloudLoggingClientI_Expecter) ListLogEntries(_a0 interface{}, _a1 interface{}) *CloudLoggingClientI_ListLogEntries_Call {
	return &CloudLoggingClientI_ListLogEntries_Call{Call: _e.mock.On("ListLogEntries", _a0, _a1)}
}

func (_c *CloudLoggingClientI_ListLogEntries_Call) Run(run func(_a0 context.Context, _a1 bigquery.ImportActivitiesFilter)) *CloudLoggingClientI_ListLogEntries_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(bigquery.ImportActivitiesFilter))
	})
	return _c
}

func (_c *CloudLoggingClientI_ListLogEntries_Call) Return(_a0 []*bigquery.Activity, _a1 error) *CloudLoggingClientI_ListLogEntries_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}
