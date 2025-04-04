// Code generated by mockery v2.11.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// Publisher is an autogenerated mock type for the Publisher type
type Publisher struct {
	mock.Mock
}

// Publish provides a mock function with given fields: ctx, _a1, key, message, metadata
func (_m *Publisher) Publish(ctx context.Context, _a1 string, key string, message interface{}, metadata map[string]interface{}) error {
	ret := _m.Called(ctx, _a1, key, message, metadata)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, interface{}, map[string]interface{}) error); ok {
		r0 = rf(ctx, _a1, key, message, metadata)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewPublisher creates a new instance of Publisher. It also registers a cleanup function to assert the mocks expectations.
func NewPublisher(t testing.TB) *Publisher {
	mock := &Publisher{}

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
