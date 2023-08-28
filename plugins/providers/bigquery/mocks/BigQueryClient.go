// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	context "context"

	bigquery "github.com/raystack/guardian/plugins/providers/bigquery"

	domain "github.com/raystack/guardian/domain"

	gobigquery "cloud.google.com/go/bigquery"

	mock "github.com/stretchr/testify/mock"
)

// BigQueryClient is an autogenerated mock type for the BigQueryClient type
type BigQueryClient struct {
	mock.Mock
}

type BigQueryClient_Expecter struct {
	mock *mock.Mock
}

func (_m *BigQueryClient) EXPECT() *BigQueryClient_Expecter {
	return &BigQueryClient_Expecter{mock: &_m.Mock}
}

// GetDatasets provides a mock function with given fields: _a0
func (_m *BigQueryClient) GetDatasets(_a0 context.Context) ([]*bigquery.Dataset, error) {
	ret := _m.Called(_a0)

	var r0 []*bigquery.Dataset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*bigquery.Dataset, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*bigquery.Dataset); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*bigquery.Dataset)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BigQueryClient_GetDatasets_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDatasets'
type BigQueryClient_GetDatasets_Call struct {
	*mock.Call
}

// GetDatasets is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *BigQueryClient_Expecter) GetDatasets(_a0 interface{}) *BigQueryClient_GetDatasets_Call {
	return &BigQueryClient_GetDatasets_Call{Call: _e.mock.On("GetDatasets", _a0)}
}

func (_c *BigQueryClient_GetDatasets_Call) Run(run func(_a0 context.Context)) *BigQueryClient_GetDatasets_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *BigQueryClient_GetDatasets_Call) Return(_a0 []*bigquery.Dataset, _a1 error) *BigQueryClient_GetDatasets_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *BigQueryClient_GetDatasets_Call) RunAndReturn(run func(context.Context) ([]*bigquery.Dataset, error)) *BigQueryClient_GetDatasets_Call {
	_c.Call.Return(run)
	return _c
}

// GetRolePermissions provides a mock function with given fields: _a0, _a1
func (_m *BigQueryClient) GetRolePermissions(_a0 context.Context, _a1 string) ([]string, error) {
	ret := _m.Called(_a0, _a1)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]string, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []string); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BigQueryClient_GetRolePermissions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRolePermissions'
type BigQueryClient_GetRolePermissions_Call struct {
	*mock.Call
}

// GetRolePermissions is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 string
func (_e *BigQueryClient_Expecter) GetRolePermissions(_a0 interface{}, _a1 interface{}) *BigQueryClient_GetRolePermissions_Call {
	return &BigQueryClient_GetRolePermissions_Call{Call: _e.mock.On("GetRolePermissions", _a0, _a1)}
}

func (_c *BigQueryClient_GetRolePermissions_Call) Run(run func(_a0 context.Context, _a1 string)) *BigQueryClient_GetRolePermissions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *BigQueryClient_GetRolePermissions_Call) Return(_a0 []string, _a1 error) *BigQueryClient_GetRolePermissions_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *BigQueryClient_GetRolePermissions_Call) RunAndReturn(run func(context.Context, string) ([]string, error)) *BigQueryClient_GetRolePermissions_Call {
	_c.Call.Return(run)
	return _c
}

// GetTables provides a mock function with given fields: ctx, datasetID
func (_m *BigQueryClient) GetTables(ctx context.Context, datasetID string) ([]*bigquery.Table, error) {
	ret := _m.Called(ctx, datasetID)

	var r0 []*bigquery.Table
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*bigquery.Table, error)); ok {
		return rf(ctx, datasetID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []*bigquery.Table); ok {
		r0 = rf(ctx, datasetID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*bigquery.Table)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, datasetID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BigQueryClient_GetTables_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTables'
type BigQueryClient_GetTables_Call struct {
	*mock.Call
}

// GetTables is a helper method to define mock.On call
//   - ctx context.Context
//   - datasetID string
func (_e *BigQueryClient_Expecter) GetTables(ctx interface{}, datasetID interface{}) *BigQueryClient_GetTables_Call {
	return &BigQueryClient_GetTables_Call{Call: _e.mock.On("GetTables", ctx, datasetID)}
}

func (_c *BigQueryClient_GetTables_Call) Run(run func(ctx context.Context, datasetID string)) *BigQueryClient_GetTables_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *BigQueryClient_GetTables_Call) Return(_a0 []*bigquery.Table, _a1 error) *BigQueryClient_GetTables_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *BigQueryClient_GetTables_Call) RunAndReturn(run func(context.Context, string) ([]*bigquery.Table, error)) *BigQueryClient_GetTables_Call {
	_c.Call.Return(run)
	return _c
}

// GrantDatasetAccess provides a mock function with given fields: ctx, d, user, role
func (_m *BigQueryClient) GrantDatasetAccess(ctx context.Context, d *bigquery.Dataset, user string, role string) error {
	ret := _m.Called(ctx, d, user, role)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *bigquery.Dataset, string, string) error); ok {
		r0 = rf(ctx, d, user, role)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// BigQueryClient_GrantDatasetAccess_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GrantDatasetAccess'
type BigQueryClient_GrantDatasetAccess_Call struct {
	*mock.Call
}

// GrantDatasetAccess is a helper method to define mock.On call
//   - ctx context.Context
//   - d *bigquery.Dataset
//   - user string
//   - role string
func (_e *BigQueryClient_Expecter) GrantDatasetAccess(ctx interface{}, d interface{}, user interface{}, role interface{}) *BigQueryClient_GrantDatasetAccess_Call {
	return &BigQueryClient_GrantDatasetAccess_Call{Call: _e.mock.On("GrantDatasetAccess", ctx, d, user, role)}
}

func (_c *BigQueryClient_GrantDatasetAccess_Call) Run(run func(ctx context.Context, d *bigquery.Dataset, user string, role string)) *BigQueryClient_GrantDatasetAccess_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*bigquery.Dataset), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *BigQueryClient_GrantDatasetAccess_Call) Return(_a0 error) *BigQueryClient_GrantDatasetAccess_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BigQueryClient_GrantDatasetAccess_Call) RunAndReturn(run func(context.Context, *bigquery.Dataset, string, string) error) *BigQueryClient_GrantDatasetAccess_Call {
	_c.Call.Return(run)
	return _c
}

// GrantTableAccess provides a mock function with given fields: ctx, t, accountType, accountID, role
func (_m *BigQueryClient) GrantTableAccess(ctx context.Context, t *bigquery.Table, accountType string, accountID string, role string) error {
	ret := _m.Called(ctx, t, accountType, accountID, role)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *bigquery.Table, string, string, string) error); ok {
		r0 = rf(ctx, t, accountType, accountID, role)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// BigQueryClient_GrantTableAccess_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GrantTableAccess'
type BigQueryClient_GrantTableAccess_Call struct {
	*mock.Call
}

// GrantTableAccess is a helper method to define mock.On call
//   - ctx context.Context
//   - t *bigquery.Table
//   - accountType string
//   - accountID string
//   - role string
func (_e *BigQueryClient_Expecter) GrantTableAccess(ctx interface{}, t interface{}, accountType interface{}, accountID interface{}, role interface{}) *BigQueryClient_GrantTableAccess_Call {
	return &BigQueryClient_GrantTableAccess_Call{Call: _e.mock.On("GrantTableAccess", ctx, t, accountType, accountID, role)}
}

func (_c *BigQueryClient_GrantTableAccess_Call) Run(run func(ctx context.Context, t *bigquery.Table, accountType string, accountID string, role string)) *BigQueryClient_GrantTableAccess_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*bigquery.Table), args[2].(string), args[3].(string), args[4].(string))
	})
	return _c
}

func (_c *BigQueryClient_GrantTableAccess_Call) Return(_a0 error) *BigQueryClient_GrantTableAccess_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BigQueryClient_GrantTableAccess_Call) RunAndReturn(run func(context.Context, *bigquery.Table, string, string, string) error) *BigQueryClient_GrantTableAccess_Call {
	_c.Call.Return(run)
	return _c
}

// ListAccess provides a mock function with given fields: ctx, resources
func (_m *BigQueryClient) ListAccess(ctx context.Context, resources []*domain.Resource) (domain.MapResourceAccess, error) {
	ret := _m.Called(ctx, resources)

	var r0 domain.MapResourceAccess
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []*domain.Resource) (domain.MapResourceAccess, error)); ok {
		return rf(ctx, resources)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []*domain.Resource) domain.MapResourceAccess); ok {
		r0 = rf(ctx, resources)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(domain.MapResourceAccess)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []*domain.Resource) error); ok {
		r1 = rf(ctx, resources)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BigQueryClient_ListAccess_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListAccess'
type BigQueryClient_ListAccess_Call struct {
	*mock.Call
}

// ListAccess is a helper method to define mock.On call
//   - ctx context.Context
//   - resources []*domain.Resource
func (_e *BigQueryClient_Expecter) ListAccess(ctx interface{}, resources interface{}) *BigQueryClient_ListAccess_Call {
	return &BigQueryClient_ListAccess_Call{Call: _e.mock.On("ListAccess", ctx, resources)}
}

func (_c *BigQueryClient_ListAccess_Call) Run(run func(ctx context.Context, resources []*domain.Resource)) *BigQueryClient_ListAccess_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]*domain.Resource))
	})
	return _c
}

func (_c *BigQueryClient_ListAccess_Call) Return(_a0 domain.MapResourceAccess, _a1 error) *BigQueryClient_ListAccess_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *BigQueryClient_ListAccess_Call) RunAndReturn(run func(context.Context, []*domain.Resource) (domain.MapResourceAccess, error)) *BigQueryClient_ListAccess_Call {
	_c.Call.Return(run)
	return _c
}

// ResolveDatasetRole provides a mock function with given fields: role
func (_m *BigQueryClient) ResolveDatasetRole(role string) (gobigquery.AccessRole, error) {
	ret := _m.Called(role)

	var r0 gobigquery.AccessRole
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (gobigquery.AccessRole, error)); ok {
		return rf(role)
	}
	if rf, ok := ret.Get(0).(func(string) gobigquery.AccessRole); ok {
		r0 = rf(role)
	} else {
		r0 = ret.Get(0).(gobigquery.AccessRole)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(role)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BigQueryClient_ResolveDatasetRole_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResolveDatasetRole'
type BigQueryClient_ResolveDatasetRole_Call struct {
	*mock.Call
}

// ResolveDatasetRole is a helper method to define mock.On call
//   - role string
func (_e *BigQueryClient_Expecter) ResolveDatasetRole(role interface{}) *BigQueryClient_ResolveDatasetRole_Call {
	return &BigQueryClient_ResolveDatasetRole_Call{Call: _e.mock.On("ResolveDatasetRole", role)}
}

func (_c *BigQueryClient_ResolveDatasetRole_Call) Run(run func(role string)) *BigQueryClient_ResolveDatasetRole_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *BigQueryClient_ResolveDatasetRole_Call) Return(_a0 gobigquery.AccessRole, _a1 error) *BigQueryClient_ResolveDatasetRole_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *BigQueryClient_ResolveDatasetRole_Call) RunAndReturn(run func(string) (gobigquery.AccessRole, error)) *BigQueryClient_ResolveDatasetRole_Call {
	_c.Call.Return(run)
	return _c
}

// RevokeDatasetAccess provides a mock function with given fields: ctx, d, user, role
func (_m *BigQueryClient) RevokeDatasetAccess(ctx context.Context, d *bigquery.Dataset, user string, role string) error {
	ret := _m.Called(ctx, d, user, role)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *bigquery.Dataset, string, string) error); ok {
		r0 = rf(ctx, d, user, role)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// BigQueryClient_RevokeDatasetAccess_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RevokeDatasetAccess'
type BigQueryClient_RevokeDatasetAccess_Call struct {
	*mock.Call
}

// RevokeDatasetAccess is a helper method to define mock.On call
//   - ctx context.Context
//   - d *bigquery.Dataset
//   - user string
//   - role string
func (_e *BigQueryClient_Expecter) RevokeDatasetAccess(ctx interface{}, d interface{}, user interface{}, role interface{}) *BigQueryClient_RevokeDatasetAccess_Call {
	return &BigQueryClient_RevokeDatasetAccess_Call{Call: _e.mock.On("RevokeDatasetAccess", ctx, d, user, role)}
}

func (_c *BigQueryClient_RevokeDatasetAccess_Call) Run(run func(ctx context.Context, d *bigquery.Dataset, user string, role string)) *BigQueryClient_RevokeDatasetAccess_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*bigquery.Dataset), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *BigQueryClient_RevokeDatasetAccess_Call) Return(_a0 error) *BigQueryClient_RevokeDatasetAccess_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BigQueryClient_RevokeDatasetAccess_Call) RunAndReturn(run func(context.Context, *bigquery.Dataset, string, string) error) *BigQueryClient_RevokeDatasetAccess_Call {
	_c.Call.Return(run)
	return _c
}

// RevokeTableAccess provides a mock function with given fields: ctx, t, accountType, accountID, role
func (_m *BigQueryClient) RevokeTableAccess(ctx context.Context, t *bigquery.Table, accountType string, accountID string, role string) error {
	ret := _m.Called(ctx, t, accountType, accountID, role)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *bigquery.Table, string, string, string) error); ok {
		r0 = rf(ctx, t, accountType, accountID, role)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// BigQueryClient_RevokeTableAccess_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RevokeTableAccess'
type BigQueryClient_RevokeTableAccess_Call struct {
	*mock.Call
}

// RevokeTableAccess is a helper method to define mock.On call
//   - ctx context.Context
//   - t *bigquery.Table
//   - accountType string
//   - accountID string
//   - role string
func (_e *BigQueryClient_Expecter) RevokeTableAccess(ctx interface{}, t interface{}, accountType interface{}, accountID interface{}, role interface{}) *BigQueryClient_RevokeTableAccess_Call {
	return &BigQueryClient_RevokeTableAccess_Call{Call: _e.mock.On("RevokeTableAccess", ctx, t, accountType, accountID, role)}
}

func (_c *BigQueryClient_RevokeTableAccess_Call) Run(run func(ctx context.Context, t *bigquery.Table, accountType string, accountID string, role string)) *BigQueryClient_RevokeTableAccess_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*bigquery.Table), args[2].(string), args[3].(string), args[4].(string))
	})
	return _c
}

func (_c *BigQueryClient_RevokeTableAccess_Call) Return(_a0 error) *BigQueryClient_RevokeTableAccess_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BigQueryClient_RevokeTableAccess_Call) RunAndReturn(run func(context.Context, *bigquery.Table, string, string, string) error) *BigQueryClient_RevokeTableAccess_Call {
	_c.Call.Return(run)
	return _c
}

// NewBigQueryClient creates a new instance of BigQueryClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBigQueryClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *BigQueryClient {
	mock := &BigQueryClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
