// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/stamp/stamp.repository.go

// Package mock_stamp is a generated GoMock package.
package mock_stamp

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	model "github.com/isd-sgcu/rpkm67-model/model"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// FindByUserId mocks base method.
func (m *MockRepository) FindByUserId(userId string, stamp *model.Stamp) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByUserId", userId, stamp)
	ret0, _ := ret[0].(error)
	return ret0
}

// FindByUserId indicates an expected call of FindByUserId.
func (mr *MockRepositoryMockRecorder) FindByUserId(userId, stamp interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByUserId", reflect.TypeOf((*MockRepository)(nil).FindByUserId), userId, stamp)
}

// StampByUserId mocks base method.
func (m *MockRepository) StampByUserId(userId string, stamp *model.Stamp) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StampByUserId", userId, stamp)
	ret0, _ := ret[0].(error)
	return ret0
}

// StampByUserId indicates an expected call of StampByUserId.
func (mr *MockRepositoryMockRecorder) StampByUserId(userId, stamp interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StampByUserId", reflect.TypeOf((*MockRepository)(nil).StampByUserId), userId, stamp)
}
