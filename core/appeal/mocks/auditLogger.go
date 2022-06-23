// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// AuditLogger is an autogenerated mock type for the auditLogger type
type AuditLogger struct {
	mock.Mock
}

// Log provides a mock function with given fields: ctx, action, data
func (_m *AuditLogger) Log(ctx context.Context, action string, data interface{}) error {
	ret := _m.Called(ctx, action, data)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, interface{}) error); ok {
		r0 = rf(ctx, action, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}