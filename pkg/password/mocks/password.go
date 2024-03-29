// Code generated by mockery v2.42.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Password is an autogenerated mock type for the Password type
type Password struct {
	mock.Mock
}

// Compare provides a mock function with given fields: pwd, hash
func (_m *Password) Compare(pwd string, hash string) bool {
	ret := _m.Called(pwd, hash)

	if len(ret) == 0 {
		panic("no return value specified for Compare")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string) bool); ok {
		r0 = rf(pwd, hash)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Hash provides a mock function with given fields: pwd
func (_m *Password) Hash(pwd string) (string, error) {
	ret := _m.Called(pwd)

	if len(ret) == 0 {
		panic("no return value specified for Hash")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(pwd)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(pwd)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(pwd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewPassword creates a new instance of Password. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPassword(t interface {
	mock.TestingT
	Cleanup(func())
}) *Password {
	mock := &Password{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
