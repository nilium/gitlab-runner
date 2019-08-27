// Code generated by mockery v1.0.0. DO NOT EDIT.

// This comment works around https://github.com/vektra/mockery/issues/155

package process

import (
	logrus "github.com/sirupsen/logrus"
	mock "github.com/stretchr/testify/mock"
)

// MockLogger is an autogenerated mock type for the Logger type
type MockLogger struct {
	mock.Mock
}

// Warn provides a mock function with given fields: args
func (_m *MockLogger) Warn(args ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, args...)
	_m.Called(_ca...)
}

// WithFields provides a mock function with given fields: fields
func (_m *MockLogger) WithFields(fields logrus.Fields) Logger {
	ret := _m.Called(fields)

	var r0 Logger
	if rf, ok := ret.Get(0).(func(logrus.Fields) Logger); ok {
		r0 = rf(fields)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Logger)
		}
	}

	return r0
}
