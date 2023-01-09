// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

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

// AddApprover provides a mock function with given fields: ctx, approvalID, email
func (_m *ApprovalService) AddApprover(ctx context.Context, approvalID string, email string) error {
	ret := _m.Called(ctx, approvalID, email)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, approvalID, email)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ApprovalService_AddApprover_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddApprover'
type ApprovalService_AddApprover_Call struct {
	*mock.Call
}

// AddApprover is a helper method to define mock.On call
//  - ctx context.Context
//  - approvalID string
//  - email string
func (_e *ApprovalService_Expecter) AddApprover(ctx interface{}, approvalID interface{}, email interface{}) *ApprovalService_AddApprover_Call {
	return &ApprovalService_AddApprover_Call{Call: _e.mock.On("AddApprover", ctx, approvalID, email)}
}

func (_c *ApprovalService_AddApprover_Call) Run(run func(ctx context.Context, approvalID string, email string)) *ApprovalService_AddApprover_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *ApprovalService_AddApprover_Call) Return(_a0 error) *ApprovalService_AddApprover_Call {
	_c.Call.Return(_a0)
	return _c
}

// DeleteApprover provides a mock function with given fields: ctx, approvalID, email
func (_m *ApprovalService) DeleteApprover(ctx context.Context, approvalID string, email string) error {
	ret := _m.Called(ctx, approvalID, email)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, approvalID, email)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ApprovalService_DeleteApprover_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteApprover'
type ApprovalService_DeleteApprover_Call struct {
	*mock.Call
}

// DeleteApprover is a helper method to define mock.On call
//  - ctx context.Context
//  - approvalID string
//  - email string
func (_e *ApprovalService_Expecter) DeleteApprover(ctx interface{}, approvalID interface{}, email interface{}) *ApprovalService_DeleteApprover_Call {
	return &ApprovalService_DeleteApprover_Call{Call: _e.mock.On("DeleteApprover", ctx, approvalID, email)}
}

func (_c *ApprovalService_DeleteApprover_Call) Run(run func(ctx context.Context, approvalID string, email string)) *ApprovalService_DeleteApprover_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *ApprovalService_DeleteApprover_Call) Return(_a0 error) *ApprovalService_DeleteApprover_Call {
	_c.Call.Return(_a0)
	return _c
}

type mockConstructorTestingTNewApprovalService interface {
	mock.TestingT
	Cleanup(func())
}

// NewApprovalService creates a new instance of ApprovalService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewApprovalService(t mockConstructorTestingTNewApprovalService) *ApprovalService {
	mock := &ApprovalService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
