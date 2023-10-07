// Code generated by mockery v2.33.3. DO NOT EDIT.

package mocks

import (
	context "context"

	bigquery "github.com/raystack/guardian/plugins/providers/bigquery"

	logging "google.golang.org/api/logging/v2"

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

// GetLogBucket provides a mock function with given fields: ctx, name
func (_m *CloudLoggingClientI) GetLogBucket(ctx context.Context, name string) (*logging.LogBucket, error) {
	ret := _m.Called(ctx, name)

	var r0 *logging.LogBucket
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*logging.LogBucket, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *logging.LogBucket); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*logging.LogBucket)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CloudLoggingClientI_GetLogBucket_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLogBucket'
type CloudLoggingClientI_GetLogBucket_Call struct {
	*mock.Call
}

// GetLogBucket is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
func (_e *CloudLoggingClientI_Expecter) GetLogBucket(ctx interface{}, name interface{}) *CloudLoggingClientI_GetLogBucket_Call {
	return &CloudLoggingClientI_GetLogBucket_Call{Call: _e.mock.On("GetLogBucket", ctx, name)}
}

func (_c *CloudLoggingClientI_GetLogBucket_Call) Run(run func(ctx context.Context, name string)) *CloudLoggingClientI_GetLogBucket_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CloudLoggingClientI_GetLogBucket_Call) Return(_a0 *logging.LogBucket, _a1 error) *CloudLoggingClientI_GetLogBucket_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CloudLoggingClientI_GetLogBucket_Call) RunAndReturn(run func(context.Context, string) (*logging.LogBucket, error)) *CloudLoggingClientI_GetLogBucket_Call {
	_c.Call.Return(run)
	return _c
}

// ListLogEntries provides a mock function with given fields: _a0, _a1, _a2
func (_m *CloudLoggingClientI) ListLogEntries(_a0 context.Context, _a1 string, _a2 int) ([]*bigquery.Activity, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 []*bigquery.Activity
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, int) ([]*bigquery.Activity, error)); ok {
		return rf(_a0, _a1, _a2)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, int) []*bigquery.Activity); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*bigquery.Activity)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, int) error); ok {
		r1 = rf(_a0, _a1, _a2)
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
//   - _a0 context.Context
//   - _a1 string
//   - _a2 int
func (_e *CloudLoggingClientI_Expecter) ListLogEntries(_a0 interface{}, _a1 interface{}, _a2 interface{}) *CloudLoggingClientI_ListLogEntries_Call {
	return &CloudLoggingClientI_ListLogEntries_Call{Call: _e.mock.On("ListLogEntries", _a0, _a1, _a2)}
}

func (_c *CloudLoggingClientI_ListLogEntries_Call) Run(run func(_a0 context.Context, _a1 string, _a2 int)) *CloudLoggingClientI_ListLogEntries_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(int))
	})
	return _c
}

func (_c *CloudLoggingClientI_ListLogEntries_Call) Return(_a0 []*bigquery.Activity, _a1 error) *CloudLoggingClientI_ListLogEntries_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CloudLoggingClientI_ListLogEntries_Call) RunAndReturn(run func(context.Context, string, int) ([]*bigquery.Activity, error)) *CloudLoggingClientI_ListLogEntries_Call {
	_c.Call.Return(run)
	return _c
}

// NewCloudLoggingClientI creates a new instance of CloudLoggingClientI. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCloudLoggingClientI(t interface {
	mock.TestingT
	Cleanup(func())
}) *CloudLoggingClientI {
	mock := &CloudLoggingClientI{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
