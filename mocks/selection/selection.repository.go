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

// CountByBaanId mocks base method.
func (m *MockRepository) CountByBaanId() (map[string]int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CountByBaanId")
	ret0, _ := ret[0].(map[string]int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CountByBaanId indicates an expected call of CountByBaanId.
func (mr *MockRepositoryMockRecorder) CountByBaanId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountByBaanId", reflect.TypeOf((*MockRepository)(nil).CountByBaanId))
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
func (m *MockRepository) Delete(groupId, baanId string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", groupId, baanId)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockRepositoryMockRecorder) Delete(groupId, baanId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockRepository)(nil).Delete), groupId, baanId)
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

// UpdateExistBaanExistOrder mocks base method.
func (m *MockRepository) UpdateExistBaanExistOrder(updateSelection *model.Selection) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateExistBaanExistOrder", updateSelection)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateExistBaanExistOrder indicates an expected call of UpdateExistBaanExistOrder.
func (mr *MockRepositoryMockRecorder) UpdateExistBaanExistOrder(updateSelection interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateExistBaanExistOrder", reflect.TypeOf((*MockRepository)(nil).UpdateExistBaanExistOrder), updateSelection)
}

// UpdateExistBaanNewOrder mocks base method.
func (m *MockRepository) UpdateExistBaanNewOrder(updateSelection *model.Selection) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateExistBaanNewOrder", updateSelection)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateExistBaanNewOrder indicates an expected call of UpdateExistBaanNewOrder.
func (mr *MockRepositoryMockRecorder) UpdateExistBaanNewOrder(updateSelection interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateExistBaanNewOrder", reflect.TypeOf((*MockRepository)(nil).UpdateExistBaanNewOrder), updateSelection)
}

// UpdateNewBaanExistOrder mocks base method.
func (m *MockRepository) UpdateNewBaanExistOrder(updateSelection *model.Selection) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNewBaanExistOrder", updateSelection)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateNewBaanExistOrder indicates an expected call of UpdateNewBaanExistOrder.
func (mr *MockRepositoryMockRecorder) UpdateNewBaanExistOrder(updateSelection interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNewBaanExistOrder", reflect.TypeOf((*MockRepository)(nil).UpdateNewBaanExistOrder), updateSelection)
}
