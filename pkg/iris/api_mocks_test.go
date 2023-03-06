// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/iris/api.go

// Package iris is a generated GoMock package.
package iris

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/leanix/leanix-k8s-connector/pkg/iris/models"
)

// MockAPI is a mock of API interface.
type MockAPI struct {
	ctrl     *gomock.Controller
	recorder *MockAPIMockRecorder
}

// MockAPIMockRecorder is the mock recorder for MockAPI.
type MockAPIMockRecorder struct {
	mock *MockAPI
}

// NewMockAPI creates a new mock instance.
func NewMockAPI(ctrl *gomock.Controller) *MockAPI {
	mock := &MockAPI{ctrl: ctrl}
	mock.recorder = &MockAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAPI) EXPECT() *MockAPIMockRecorder {
	return m.recorder
}

// GetConfiguration mocks base method.
func (m *MockAPI) GetConfiguration(configurationName string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConfiguration", configurationName)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetConfiguration indicates an expected call of GetConfiguration.
func (mr *MockAPIMockRecorder) GetConfiguration(configurationName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConfiguration", reflect.TypeOf((*MockAPI)(nil).GetConfiguration), configurationName)
}

// GetScanResults mocks base method.
func (m *MockAPI) GetScanResults(configurationId string) ([]models.DiscoveryEvent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetScanResults", configurationId)
	ret0, _ := ret[0].([]models.DiscoveryEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetScanResults indicates an expected call of GetScanResults.
func (mr *MockAPIMockRecorder) GetScanResults(configurationId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetScanResults", reflect.TypeOf((*MockAPI)(nil).GetScanResults), configurationId)
}

// PostEcstResults mocks base method.
func (m *MockAPI) PostEcstResults(ecstResults []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PostEcstResults", ecstResults)
	ret0, _ := ret[0].(error)
	return ret0
}

// PostEcstResults indicates an expected call of PostEcstResults.
func (mr *MockAPIMockRecorder) PostEcstResults(ecstResults interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostEcstResults", reflect.TypeOf((*MockAPI)(nil).PostEcstResults), ecstResults)
}

// PostResults mocks base method.
func (m *MockAPI) PostResults(results []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PostResults", results)
	ret0, _ := ret[0].(error)
	return ret0
}

// PostResults indicates an expected call of PostResults.
func (mr *MockAPIMockRecorder) PostResults(results interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostResults", reflect.TypeOf((*MockAPI)(nil).PostResults), results)
}

// PostStatus mocks base method.
func (m *MockAPI) PostStatus(status []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PostStatus", status)
	ret0, _ := ret[0].(error)
	return ret0
}

// PostStatus indicates an expected call of PostStatus.
func (mr *MockAPIMockRecorder) PostStatus(status interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostStatus", reflect.TypeOf((*MockAPI)(nil).PostStatus), status)
}
