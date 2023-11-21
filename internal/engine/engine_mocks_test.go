// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/controlplaneio/netassert/internal/engine (interfaces: NetAssertTestRunner)

// Package engine is a generated GoMock package.
package engine

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "go.uber.org/mock/gomock"
	v1 "k8s.io/api/core/v1"
)

// MockNetAssertTestRunner is a mock of NetAssertTestRunner interface.
type MockNetAssertTestRunner struct {
	ctrl     *gomock.Controller
	recorder *MockNetAssertTestRunnerMockRecorder
}

// MockNetAssertTestRunnerMockRecorder is the mock recorder for MockNetAssertTestRunner.
type MockNetAssertTestRunnerMockRecorder struct {
	mock *MockNetAssertTestRunner
}

// NewMockNetAssertTestRunner creates a new mock instance.
func NewMockNetAssertTestRunner(ctrl *gomock.Controller) *MockNetAssertTestRunner {
	mock := &MockNetAssertTestRunner{ctrl: ctrl}
	mock.recorder = &MockNetAssertTestRunnerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNetAssertTestRunner) EXPECT() *MockNetAssertTestRunnerMockRecorder {
	return m.recorder
}

// BuildEphemeralScannerContainer mocks base method.
func (m *MockNetAssertTestRunner) BuildEphemeralScannerContainer(arg0, arg1, arg2, arg3, arg4, arg5 string, arg6 int) (*v1.EphemeralContainer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuildEphemeralScannerContainer", arg0, arg1, arg2, arg3, arg4, arg5, arg6)
	ret0, _ := ret[0].(*v1.EphemeralContainer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BuildEphemeralScannerContainer indicates an expected call of BuildEphemeralScannerContainer.
func (mr *MockNetAssertTestRunnerMockRecorder) BuildEphemeralScannerContainer(arg0, arg1, arg2, arg3, arg4, arg5, arg6 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildEphemeralScannerContainer", reflect.TypeOf((*MockNetAssertTestRunner)(nil).BuildEphemeralScannerContainer), arg0, arg1, arg2, arg3, arg4, arg5, arg6)
}

// BuildEphemeralSnifferContainer mocks base method.
func (m *MockNetAssertTestRunner) BuildEphemeralSnifferContainer(arg0, arg1, arg2 string, arg3 int, arg4 string, arg5 int, arg6 string, arg7 int) (*v1.EphemeralContainer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuildEphemeralSnifferContainer", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	ret0, _ := ret[0].(*v1.EphemeralContainer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BuildEphemeralSnifferContainer indicates an expected call of BuildEphemeralSnifferContainer.
func (mr *MockNetAssertTestRunnerMockRecorder) BuildEphemeralSnifferContainer(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildEphemeralSnifferContainer", reflect.TypeOf((*MockNetAssertTestRunner)(nil).BuildEphemeralSnifferContainer), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
}

// GetExitStatusOfEphemeralContainer mocks base method.
func (m *MockNetAssertTestRunner) GetExitStatusOfEphemeralContainer(arg0 context.Context, arg1 string, arg2 time.Duration, arg3, arg4 string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExitStatusOfEphemeralContainer", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExitStatusOfEphemeralContainer indicates an expected call of GetExitStatusOfEphemeralContainer.
func (mr *MockNetAssertTestRunnerMockRecorder) GetExitStatusOfEphemeralContainer(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExitStatusOfEphemeralContainer", reflect.TypeOf((*MockNetAssertTestRunner)(nil).GetExitStatusOfEphemeralContainer), arg0, arg1, arg2, arg3, arg4)
}

// GetPod mocks base method.
func (m *MockNetAssertTestRunner) GetPod(arg0 context.Context, arg1, arg2 string) (*v1.Pod, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPod", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.Pod)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPod indicates an expected call of GetPod.
func (mr *MockNetAssertTestRunnerMockRecorder) GetPod(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPod", reflect.TypeOf((*MockNetAssertTestRunner)(nil).GetPod), arg0, arg1, arg2)
}

// GetPodInDaemonSet mocks base method.
func (m *MockNetAssertTestRunner) GetPodInDaemonSet(arg0 context.Context, arg1, arg2 string) (*v1.Pod, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPodInDaemonSet", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.Pod)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPodInDaemonSet indicates an expected call of GetPodInDaemonSet.
func (mr *MockNetAssertTestRunnerMockRecorder) GetPodInDaemonSet(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPodInDaemonSet", reflect.TypeOf((*MockNetAssertTestRunner)(nil).GetPodInDaemonSet), arg0, arg1, arg2)
}

// GetPodInDeployment mocks base method.
func (m *MockNetAssertTestRunner) GetPodInDeployment(arg0 context.Context, arg1, arg2 string) (*v1.Pod, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPodInDeployment", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.Pod)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPodInDeployment indicates an expected call of GetPodInDeployment.
func (mr *MockNetAssertTestRunnerMockRecorder) GetPodInDeployment(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPodInDeployment", reflect.TypeOf((*MockNetAssertTestRunner)(nil).GetPodInDeployment), arg0, arg1, arg2)
}

// GetPodInStatefulSet mocks base method.
func (m *MockNetAssertTestRunner) GetPodInStatefulSet(arg0 context.Context, arg1, arg2 string) (*v1.Pod, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPodInStatefulSet", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.Pod)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPodInStatefulSet indicates an expected call of GetPodInStatefulSet.
func (mr *MockNetAssertTestRunnerMockRecorder) GetPodInStatefulSet(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPodInStatefulSet", reflect.TypeOf((*MockNetAssertTestRunner)(nil).GetPodInStatefulSet), arg0, arg1, arg2)
}

// LaunchEphemeralContainerInPod mocks base method.
func (m *MockNetAssertTestRunner) LaunchEphemeralContainerInPod(arg0 context.Context, arg1 *v1.Pod, arg2 *v1.EphemeralContainer) (*v1.Pod, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LaunchEphemeralContainerInPod", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.Pod)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// LaunchEphemeralContainerInPod indicates an expected call of LaunchEphemeralContainerInPod.
func (mr *MockNetAssertTestRunnerMockRecorder) LaunchEphemeralContainerInPod(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LaunchEphemeralContainerInPod", reflect.TypeOf((*MockNetAssertTestRunner)(nil).LaunchEphemeralContainerInPod), arg0, arg1, arg2)
}
