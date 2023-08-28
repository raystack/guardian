// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/raystack/guardian/domain"
	gcs "github.com/raystack/guardian/plugins/providers/gcs"

	iam "cloud.google.com/go/iam"

	mock "github.com/stretchr/testify/mock"
)

// GCSClient is an autogenerated mock type for the GCSClient type
type GCSClient struct {
	mock.Mock
}

type GCSClient_Expecter struct {
	mock *mock.Mock
}

func (_m *GCSClient) EXPECT() *GCSClient_Expecter {
	return &GCSClient_Expecter{mock: &_m.Mock}
}

// GetBuckets provides a mock function with given fields: _a0
func (_m *GCSClient) GetBuckets(_a0 context.Context) ([]*gcs.Bucket, error) {
	ret := _m.Called(_a0)

	var r0 []*gcs.Bucket
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*gcs.Bucket, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*gcs.Bucket); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*gcs.Bucket)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GCSClient_GetBuckets_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetBuckets'
type GCSClient_GetBuckets_Call struct {
	*mock.Call
}

// GetBuckets is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *GCSClient_Expecter) GetBuckets(_a0 interface{}) *GCSClient_GetBuckets_Call {
	return &GCSClient_GetBuckets_Call{Call: _e.mock.On("GetBuckets", _a0)}
}

func (_c *GCSClient_GetBuckets_Call) Run(run func(_a0 context.Context)) *GCSClient_GetBuckets_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *GCSClient_GetBuckets_Call) Return(_a0 []*gcs.Bucket, _a1 error) *GCSClient_GetBuckets_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *GCSClient_GetBuckets_Call) RunAndReturn(run func(context.Context) ([]*gcs.Bucket, error)) *GCSClient_GetBuckets_Call {
	_c.Call.Return(run)
	return _c
}

// GrantBucketAccess provides a mock function with given fields: ctx, b, identity, roleName
func (_m *GCSClient) GrantBucketAccess(ctx context.Context, b gcs.Bucket, identity string, roleName iam.RoleName) error {
	ret := _m.Called(ctx, b, identity, roleName)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, gcs.Bucket, string, iam.RoleName) error); ok {
		r0 = rf(ctx, b, identity, roleName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GCSClient_GrantBucketAccess_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GrantBucketAccess'
type GCSClient_GrantBucketAccess_Call struct {
	*mock.Call
}

// GrantBucketAccess is a helper method to define mock.On call
//   - ctx context.Context
//   - b gcs.Bucket
//   - identity string
//   - roleName iam.RoleName
func (_e *GCSClient_Expecter) GrantBucketAccess(ctx interface{}, b interface{}, identity interface{}, roleName interface{}) *GCSClient_GrantBucketAccess_Call {
	return &GCSClient_GrantBucketAccess_Call{Call: _e.mock.On("GrantBucketAccess", ctx, b, identity, roleName)}
}

func (_c *GCSClient_GrantBucketAccess_Call) Run(run func(ctx context.Context, b gcs.Bucket, identity string, roleName iam.RoleName)) *GCSClient_GrantBucketAccess_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(gcs.Bucket), args[2].(string), args[3].(iam.RoleName))
	})
	return _c
}

func (_c *GCSClient_GrantBucketAccess_Call) Return(_a0 error) *GCSClient_GrantBucketAccess_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GCSClient_GrantBucketAccess_Call) RunAndReturn(run func(context.Context, gcs.Bucket, string, iam.RoleName) error) *GCSClient_GrantBucketAccess_Call {
	_c.Call.Return(run)
	return _c
}

// ListAccess provides a mock function with given fields: _a0, _a1
func (_m *GCSClient) ListAccess(_a0 context.Context, _a1 []*domain.Resource) (domain.MapResourceAccess, error) {
	ret := _m.Called(_a0, _a1)

	var r0 domain.MapResourceAccess
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []*domain.Resource) (domain.MapResourceAccess, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []*domain.Resource) domain.MapResourceAccess); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(domain.MapResourceAccess)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []*domain.Resource) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GCSClient_ListAccess_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListAccess'
type GCSClient_ListAccess_Call struct {
	*mock.Call
}

// ListAccess is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 []*domain.Resource
func (_e *GCSClient_Expecter) ListAccess(_a0 interface{}, _a1 interface{}) *GCSClient_ListAccess_Call {
	return &GCSClient_ListAccess_Call{Call: _e.mock.On("ListAccess", _a0, _a1)}
}

func (_c *GCSClient_ListAccess_Call) Run(run func(_a0 context.Context, _a1 []*domain.Resource)) *GCSClient_ListAccess_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]*domain.Resource))
	})
	return _c
}

func (_c *GCSClient_ListAccess_Call) Return(_a0 domain.MapResourceAccess, _a1 error) *GCSClient_ListAccess_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *GCSClient_ListAccess_Call) RunAndReturn(run func(context.Context, []*domain.Resource) (domain.MapResourceAccess, error)) *GCSClient_ListAccess_Call {
	_c.Call.Return(run)
	return _c
}

// RevokeBucketAccess provides a mock function with given fields: ctx, b, identity, roleName
func (_m *GCSClient) RevokeBucketAccess(ctx context.Context, b gcs.Bucket, identity string, roleName iam.RoleName) error {
	ret := _m.Called(ctx, b, identity, roleName)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, gcs.Bucket, string, iam.RoleName) error); ok {
		r0 = rf(ctx, b, identity, roleName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GCSClient_RevokeBucketAccess_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RevokeBucketAccess'
type GCSClient_RevokeBucketAccess_Call struct {
	*mock.Call
}

// RevokeBucketAccess is a helper method to define mock.On call
//   - ctx context.Context
//   - b gcs.Bucket
//   - identity string
//   - roleName iam.RoleName
func (_e *GCSClient_Expecter) RevokeBucketAccess(ctx interface{}, b interface{}, identity interface{}, roleName interface{}) *GCSClient_RevokeBucketAccess_Call {
	return &GCSClient_RevokeBucketAccess_Call{Call: _e.mock.On("RevokeBucketAccess", ctx, b, identity, roleName)}
}

func (_c *GCSClient_RevokeBucketAccess_Call) Run(run func(ctx context.Context, b gcs.Bucket, identity string, roleName iam.RoleName)) *GCSClient_RevokeBucketAccess_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(gcs.Bucket), args[2].(string), args[3].(iam.RoleName))
	})
	return _c
}

func (_c *GCSClient_RevokeBucketAccess_Call) Return(_a0 error) *GCSClient_RevokeBucketAccess_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GCSClient_RevokeBucketAccess_Call) RunAndReturn(run func(context.Context, gcs.Bucket, string, iam.RoleName) error) *GCSClient_RevokeBucketAccess_Call {
	_c.Call.Return(run)
	return _c
}

// NewGCSClient creates a new instance of GCSClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGCSClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *GCSClient {
	mock := &GCSClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
