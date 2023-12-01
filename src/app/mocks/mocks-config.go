// Code generated by MockGen. DO NOT EDIT.
// Source: src/app/proxy/sampler-config.go
//
// Generated by this command:
//
//	mockgen -destination src/app/mocks/mocks-config.go -package mocks -source src/app/proxy/sampler-config.go
//
// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	configuration "github.com/snivilised/cobrass/src/assistant/configuration"
	clif "github.com/snivilised/cobrass/src/clif"
	proxy "github.com/snivilised/pixa/src/app/proxy"
	gomock "go.uber.org/mock/gomock"
)

// MockProfilesConfig is a mock of ProfilesConfig interface.
type MockProfilesConfig struct {
	ctrl     *gomock.Controller
	recorder *MockProfilesConfigMockRecorder
}

// MockProfilesConfigMockRecorder is the mock recorder for MockProfilesConfig.
type MockProfilesConfigMockRecorder struct {
	mock *MockProfilesConfig
}

// NewMockProfilesConfig creates a new mock instance.
func NewMockProfilesConfig(ctrl *gomock.Controller) *MockProfilesConfig {
	mock := &MockProfilesConfig{ctrl: ctrl}
	mock.recorder = &MockProfilesConfigMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProfilesConfig) EXPECT() *MockProfilesConfigMockRecorder {
	return m.recorder
}

// Profile mocks base method.
func (m *MockProfilesConfig) Profile(name string) (clif.ChangedFlagsMap, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Profile", name)
	ret0, _ := ret[0].(clif.ChangedFlagsMap)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// Profile indicates an expected call of Profile.
func (mr *MockProfilesConfigMockRecorder) Profile(name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Profile", reflect.TypeOf((*MockProfilesConfig)(nil).Profile), name)
}

// MockProfilesConfigReader is a mock of ProfilesConfigReader interface.
type MockProfilesConfigReader struct {
	ctrl     *gomock.Controller
	recorder *MockProfilesConfigReaderMockRecorder
}

// MockProfilesConfigReaderMockRecorder is the mock recorder for MockProfilesConfigReader.
type MockProfilesConfigReaderMockRecorder struct {
	mock *MockProfilesConfigReader
}

// NewMockProfilesConfigReader creates a new mock instance.
func NewMockProfilesConfigReader(ctrl *gomock.Controller) *MockProfilesConfigReader {
	mock := &MockProfilesConfigReader{ctrl: ctrl}
	mock.recorder = &MockProfilesConfigReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProfilesConfigReader) EXPECT() *MockProfilesConfigReaderMockRecorder {
	return m.recorder
}

// Read mocks base method.
func (m *MockProfilesConfigReader) Read(arg0 configuration.ViperConfig) (proxy.ProfilesConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", arg0)
	ret0, _ := ret[0].(proxy.ProfilesConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockProfilesConfigReaderMockRecorder) Read(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockProfilesConfigReader)(nil).Read), arg0)
}

// MockSamplerConfig is a mock of SamplerConfig interface.
type MockSamplerConfig struct {
	ctrl     *gomock.Controller
	recorder *MockSamplerConfigMockRecorder
}

// MockSamplerConfigMockRecorder is the mock recorder for MockSamplerConfig.
type MockSamplerConfigMockRecorder struct {
	mock *MockSamplerConfig
}

// NewMockSamplerConfig creates a new mock instance.
func NewMockSamplerConfig(ctrl *gomock.Controller) *MockSamplerConfig {
	mock := &MockSamplerConfig{ctrl: ctrl}
	mock.recorder = &MockSamplerConfigMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSamplerConfig) EXPECT() *MockSamplerConfigMockRecorder {
	return m.recorder
}

// NoFiles mocks base method.
func (m *MockSamplerConfig) NoFiles() uint {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NoFiles")
	ret0, _ := ret[0].(uint)
	return ret0
}

// NoFiles indicates an expected call of NoFiles.
func (mr *MockSamplerConfigMockRecorder) NoFiles() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NoFiles", reflect.TypeOf((*MockSamplerConfig)(nil).NoFiles))
}

// NoFolders mocks base method.
func (m *MockSamplerConfig) NoFolders() uint {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NoFolders")
	ret0, _ := ret[0].(uint)
	return ret0
}

// NoFolders indicates an expected call of NoFolders.
func (mr *MockSamplerConfigMockRecorder) NoFolders() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NoFolders", reflect.TypeOf((*MockSamplerConfig)(nil).NoFolders))
}

// Scheme mocks base method.
func (m *MockSamplerConfig) Scheme(name string) (proxy.MsSchemeConfig, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Scheme", name)
	ret0, _ := ret[0].(proxy.MsSchemeConfig)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// Scheme indicates an expected call of Scheme.
func (mr *MockSamplerConfigMockRecorder) Scheme(name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Scheme", reflect.TypeOf((*MockSamplerConfig)(nil).Scheme), name)
}

// Validate mocks base method.
func (m *MockSamplerConfig) Validate(name string, profiles proxy.ProfilesConfig) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validate", name, profiles)
	ret0, _ := ret[0].(error)
	return ret0
}

// Validate indicates an expected call of Validate.
func (mr *MockSamplerConfigMockRecorder) Validate(name, profiles any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validate", reflect.TypeOf((*MockSamplerConfig)(nil).Validate), name, profiles)
}

// MockSamplerConfigReader is a mock of SamplerConfigReader interface.
type MockSamplerConfigReader struct {
	ctrl     *gomock.Controller
	recorder *MockSamplerConfigReaderMockRecorder
}

// MockSamplerConfigReaderMockRecorder is the mock recorder for MockSamplerConfigReader.
type MockSamplerConfigReaderMockRecorder struct {
	mock *MockSamplerConfigReader
}

// NewMockSamplerConfigReader creates a new mock instance.
func NewMockSamplerConfigReader(ctrl *gomock.Controller) *MockSamplerConfigReader {
	mock := &MockSamplerConfigReader{ctrl: ctrl}
	mock.recorder = &MockSamplerConfigReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSamplerConfigReader) EXPECT() *MockSamplerConfigReaderMockRecorder {
	return m.recorder
}

// Read mocks base method.
func (m *MockSamplerConfigReader) Read(arg0 configuration.ViperConfig) (proxy.SamplerConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", arg0)
	ret0, _ := ret[0].(proxy.SamplerConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockSamplerConfigReaderMockRecorder) Read(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockSamplerConfigReader)(nil).Read), arg0)
}
