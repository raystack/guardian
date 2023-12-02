// Code generated by mockery v2.38.0. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/raystack/guardian/domain"
	mock "github.com/stretchr/testify/mock"
)

// ApprovalService is an autogenerated mock type for the approvalService type
type ApprovalService struct {
	mock.Mock
}

type ApprovalService_Expecter struct {
	mock *mock.Mock
}

func (_m *ApprovalService) EXPECT() *ApprovalService_Expecter {
	return &ApprovalService_Expecter{mock: &_m.Mock}
}

// BulkInsert provides a mock function with given fields: _a0, _a1
func (_m *ApprovalService) BulkInsert(_a0 context.Context, _a1 []*domain.Approval) error {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for BulkInsert")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []*domain.Approval) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ApprovalService_BulkInsert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BulkInsert'
type ApprovalService_BulkInsert_Call struct {
	*mock.Call
}

// BulkInsert is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 []*domain.Approval
func (_e *ApprovalService_Expecter) BulkInsert(_a0 interface{}, _a1 interface{}) *ApprovalService_BulkInsert_Call {
	return &ApprovalService_BulkInsert_Call{Call: _e.mock.On("BulkInsert", _a0, _a1)}
}

func (_c *ApprovalService_BulkInsert_Call) Run(run func(_a0 context.Context, _a1 []*domain.Approval)) *ApprovalService_BulkInsert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]*domain.Approval))
	})
	return _c
}

func (_c *ApprovalService_BulkInsert_Call) Return(_a0 error) *ApprovalService_BulkInsert_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ApprovalService_BulkInsert_Call) RunAndReturn(run func(context.Context, []*domain.Approval) error) *ApprovalService_BulkInsert_Call {
	_c.Call.Return(run)
	return _c
}

// GetApprovalsTotalCount provides a mock function with given fields: _a0, _a1
func (_m *ApprovalService) GetApprovalsTotalCount(_a0 context.Context, _a1 *domain.ListApprovalsFilter) (int64, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for GetApprovalsTotalCount")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.ListApprovalsFilter) (int64, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *domain.ListApprovalsFilter) int64); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *domain.ListApprovalsFilter) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ApprovalService_GetApprovalsTotalCount_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetApprovalsTotalCount'
type ApprovalService_GetApprovalsTotalCount_Call struct {
	*mock.Call
}

// GetApprovalsTotalCount is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *domain.ListApprovalsFilter
func (_e *ApprovalService_Expecter) GetApprovalsTotalCount(_a0 interface{}, _a1 interface{}) *ApprovalService_GetApprovalsTotalCount_Call {
	return &ApprovalService_GetApprovalsTotalCount_Call{Call: _e.mock.On("GetApprovalsTotalCount", _a0, _a1)}
}

func (_c *ApprovalService_GetApprovalsTotalCount_Call) Run(run func(_a0 context.Context, _a1 *domain.ListApprovalsFilter)) *ApprovalService_GetApprovalsTotalCount_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*domain.ListApprovalsFilter))
	})
	return _c
}

func (_c *ApprovalService_GetApprovalsTotalCount_Call) Return(_a0 int64, _a1 error) *ApprovalService_GetApprovalsTotalCount_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ApprovalService_GetApprovalsTotalCount_Call) RunAndReturn(run func(context.Context, *domain.ListApprovalsFilter) (int64, error)) *ApprovalService_GetApprovalsTotalCount_Call {
	_c.Call.Return(run)
	return _c
}

// ListApprovals provides a mock function with given fields: _a0, _a1
func (_m *ApprovalService) ListApprovals(_a0 context.Context, _a1 *domain.ListApprovalsFilter) ([]*domain.Approval, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for ListApprovals")
	}

	var r0 []*domain.Approval
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.ListApprovalsFilter) ([]*domain.Approval, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *domain.ListApprovalsFilter) []*domain.Approval); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Approval)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *domain.ListApprovalsFilter) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ApprovalService_ListApprovals_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListApprovals'
type ApprovalService_ListApprovals_Call struct {
	*mock.Call
}

// ListApprovals is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *domain.ListApprovalsFilter
func (_e *ApprovalService_Expecter) ListApprovals(_a0 interface{}, _a1 interface{}) *ApprovalService_ListApprovals_Call {
	return &ApprovalService_ListApprovals_Call{Call: _e.mock.On("ListApprovals", _a0, _a1)}
}

func (_c *ApprovalService_ListApprovals_Call) Run(run func(_a0 context.Context, _a1 *domain.ListApprovalsFilter)) *ApprovalService_ListApprovals_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*domain.ListApprovalsFilter))
	})
	return _c
}

func (_c *ApprovalService_ListApprovals_Call) Return(_a0 []*domain.Approval, _a1 error) *ApprovalService_ListApprovals_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ApprovalService_ListApprovals_Call) RunAndReturn(run func(context.Context, *domain.ListApprovalsFilter) ([]*domain.Approval, error)) *ApprovalService_ListApprovals_Call {
	_c.Call.Return(run)
	return _c
}

// NewApprovalService creates a new instance of ApprovalService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewApprovalService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ApprovalService {
	mock := &ApprovalService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
