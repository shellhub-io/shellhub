// Code generated by mockery v2.20.0. DO NOT EDIT.

package mocks

import (
	context "context"

	revdial "github.com/shellhub-io/shellhub/pkg/revdial"
	mock "github.com/stretchr/testify/mock"
)

// IReverser is an autogenerated mock type for the IReverser type
type IReverser struct {
	mock.Mock
}

// Auth provides a mock function with given fields: ctx, token, connPath
func (_m *IReverser) Auth(ctx context.Context, token string, connPath string) error {
	ret := _m.Called(ctx, token, connPath)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, token, connPath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewListener provides a mock function with given fields:
func (_m *IReverser) NewListener() (*revdial.Listener, error) {
	ret := _m.Called()

	var r0 *revdial.Listener
	var r1 error
	if rf, ok := ret.Get(0).(func() (*revdial.Listener, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *revdial.Listener); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*revdial.Listener)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewIReverser interface {
	mock.TestingT
	Cleanup(func())
}

// NewIReverser creates a new instance of IReverser. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIReverser(t mockConstructorTestingTNewIReverser) *IReverser {
	mock := &IReverser{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
