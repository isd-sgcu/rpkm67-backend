// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/selection/selection.repository.go

// Package mock_selection is a generated GoMock package.
package mock_selection

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

// CountGroupByBaanId mocks base method.
func (m *MockRepository) CountGroupByBaanId() (map[string]int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CountGroupByBaanId")
	ret0, _ := ret[0].(map[string]int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CountGroupByBaanId indicates an expected call of CountGroupByBaanId.
func (mr *MockRepositoryMockRecorder) CountGroupByBaanId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountGroupByBaanId", reflect.TypeOf((*MockRepository)(nil).CountGroupByBaanId))
}

// Create mocks base method.
func (m *MockRepository) Create(user *model.Selection) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", user)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockRepositoryMockRecorder) Create(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockRepository)(nil).Create), user)
}

// Delete mocks base method.
func (m *MockRepository) Delete(id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockRepositoryMockRecorder) Delete(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockRepository)(nil).Delete), id)
}

// FindByGroupId mocks base method.
func (m *MockRepository) FindByGroupId(groupId string, selections *[]model.Selection) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByGroupId", groupId, selections)
	ret0, _ := ret[0].(error)
	return ret0
}

// FindByGroupId indicates an expected call of FindByGroupId.
func (mr *MockRepositoryMockRecorder) FindByGroupId(groupId, selections interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByGroupId", reflect.TypeOf((*MockRepository)(nil).FindByGroupId), groupId, selections)
}
