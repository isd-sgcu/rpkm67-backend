// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/pin/pin.utils.go

// Package mock_pin is a generated GoMock package.
package mock_pin

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockUtils is a mock of Utils interface.
type MockUtils struct {
	ctrl     *gomock.Controller
	recorder *MockUtilsMockRecorder
}

// MockUtilsMockRecorder is the mock recorder for MockUtils.
type MockUtilsMockRecorder struct {
	mock *MockUtils
}

// NewMockUtils creates a new mock instance.
func NewMockUtils(ctrl *gomock.Controller) *MockUtils {
	mock := &MockUtils{ctrl: ctrl}
	mock.recorder = &MockUtilsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUtils) EXPECT() *MockUtilsMockRecorder {
	return m.recorder
}

// GeneratePIN mocks base method.
func (m *MockUtils) GeneratePIN() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GeneratePIN")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GeneratePIN indicates an expected call of GeneratePIN.
func (mr *MockUtilsMockRecorder) GeneratePIN() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GeneratePIN", reflect.TypeOf((*MockUtils)(nil).GeneratePIN))
}
