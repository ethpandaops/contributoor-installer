// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ethpandaops/contributoor-installer/internal/sidecar (interfaces: DockerSidecar)
//
// Generated by this command:
//
//	mockgen -package mock -destination mock/docker.mock.go github.com/ethpandaops/contributoor-installer/internal/sidecar DockerSidecar
//

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockDockerSidecar is a mock of DockerSidecar interface.
type MockDockerSidecar struct {
	ctrl     *gomock.Controller
	recorder *MockDockerSidecarMockRecorder
	isgomock struct{}
}

// MockDockerSidecarMockRecorder is the mock recorder for MockDockerSidecar.
type MockDockerSidecarMockRecorder struct {
	mock *MockDockerSidecar
}

// NewMockDockerSidecar creates a new mock instance.
func NewMockDockerSidecar(ctrl *gomock.Controller) *MockDockerSidecar {
	mock := &MockDockerSidecar{ctrl: ctrl}
	mock.recorder = &MockDockerSidecarMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDockerSidecar) EXPECT() *MockDockerSidecarMockRecorder {
	return m.recorder
}

// GetComposeEnv mocks base method.
func (m *MockDockerSidecar) GetComposeEnv() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetComposeEnv")
	ret0, _ := ret[0].([]string)
	return ret0
}

// GetComposeEnv indicates an expected call of GetComposeEnv.
func (mr *MockDockerSidecarMockRecorder) GetComposeEnv() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetComposeEnv", reflect.TypeOf((*MockDockerSidecar)(nil).GetComposeEnv))
}

// IsRunning mocks base method.
func (m *MockDockerSidecar) IsRunning() (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsRunning")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsRunning indicates an expected call of IsRunning.
func (mr *MockDockerSidecarMockRecorder) IsRunning() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsRunning", reflect.TypeOf((*MockDockerSidecar)(nil).IsRunning))
}

// Logs mocks base method.
func (m *MockDockerSidecar) Logs(tailLines int, follow bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Logs", tailLines, follow)
	ret0, _ := ret[0].(error)
	return ret0
}

// Logs indicates an expected call of Logs.
func (mr *MockDockerSidecarMockRecorder) Logs(tailLines, follow any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Logs", reflect.TypeOf((*MockDockerSidecar)(nil).Logs), tailLines, follow)
}

// Start mocks base method.
func (m *MockDockerSidecar) Start() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockDockerSidecarMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockDockerSidecar)(nil).Start))
}

// Status mocks base method.
func (m *MockDockerSidecar) Status() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Status")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Status indicates an expected call of Status.
func (mr *MockDockerSidecarMockRecorder) Status() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Status", reflect.TypeOf((*MockDockerSidecar)(nil).Status))
}

// Stop mocks base method.
func (m *MockDockerSidecar) Stop() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop")
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockDockerSidecarMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockDockerSidecar)(nil).Stop))
}

// Update mocks base method.
func (m *MockDockerSidecar) Update() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update")
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockDockerSidecarMockRecorder) Update() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockDockerSidecar)(nil).Update))
}

// Version mocks base method.
func (m *MockDockerSidecar) Version() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Version")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Version indicates an expected call of Version.
func (mr *MockDockerSidecarMockRecorder) Version() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Version", reflect.TypeOf((*MockDockerSidecar)(nil).Version))
}
