// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/dappley/go-dappley/core (interfaces: Consensus,NetService,BlockPoolInterface)

// Package mocks is a generated GoMock package.
package mocks

import (
	core "github.com/dappley/go-dappley/core"
	gomock "github.com/golang/mock/gomock"
	go_libp2p_peer "github.com/libp2p/go-libp2p-peer"
	reflect "reflect"
)

// MockConsensus is a mock of Consensus interface
type MockConsensus struct {
	ctrl     *gomock.Controller
	recorder *MockConsensusMockRecorder
}

// MockConsensusMockRecorder is the mock recorder for MockConsensus
type MockConsensusMockRecorder struct {
	mock *MockConsensus
}

// NewMockConsensus creates a new mock instance
func NewMockConsensus(ctrl *gomock.Controller) *MockConsensus {
	mock := &MockConsensus{ctrl: ctrl}
	mock.recorder = &MockConsensusMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConsensus) EXPECT() *MockConsensusMockRecorder {
	return m.recorder
}

// AddProducer mocks base method
func (m *MockConsensus) AddProducer(arg0 string) error {
	ret := m.ctrl.Call(m, "AddProducer", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddProducer indicates an expected call of AddProducer
func (mr *MockConsensusMockRecorder) AddProducer(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddProducer", reflect.TypeOf((*MockConsensus)(nil).AddProducer), arg0)
}

// GetProducers mocks base method
func (m *MockConsensus) GetProducers() []string {
	ret := m.ctrl.Call(m, "GetProducers")
	ret0, _ := ret[0].([]string)
	return ret0
}

// GetProducers indicates an expected call of GetProducers
func (mr *MockConsensusMockRecorder) GetProducers() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProducers", reflect.TypeOf((*MockConsensus)(nil).GetProducers))
}

// IsProducingBlock mocks base method
func (m *MockConsensus) IsProducingBlock() bool {
	ret := m.ctrl.Call(m, "IsProducingBlock")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsProducingBlock indicates an expected call of IsProducingBlock
func (mr *MockConsensusMockRecorder) IsProducingBlock() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsProducingBlock", reflect.TypeOf((*MockConsensus)(nil).IsProducingBlock))
}

// SetKey mocks base method
func (m *MockConsensus) SetKey(arg0 string) {
	m.ctrl.Call(m, "SetKey", arg0)
}

// SetKey indicates an expected call of SetKey
func (mr *MockConsensusMockRecorder) SetKey(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetKey", reflect.TypeOf((*MockConsensus)(nil).SetKey), arg0)
}

// Setup mocks base method
func (m *MockConsensus) Setup(arg0 core.NetService, arg1 string) {
	m.ctrl.Call(m, "Setup", arg0, arg1)
}

// Setup indicates an expected call of Setup
func (mr *MockConsensusMockRecorder) Setup(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Setup", reflect.TypeOf((*MockConsensus)(nil).Setup), arg0, arg1)
}

// Start mocks base method
func (m *MockConsensus) Start() {
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start
func (mr *MockConsensusMockRecorder) Start() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockConsensus)(nil).Start))
}

// Stop mocks base method
func (m *MockConsensus) Stop() {
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop
func (mr *MockConsensusMockRecorder) Stop() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockConsensus)(nil).Stop))
}

// Validate mocks base method
func (m *MockConsensus) Validate(arg0 *core.Block) bool {
	ret := m.ctrl.Call(m, "Validate", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Validate indicates an expected call of Validate
func (mr *MockConsensusMockRecorder) Validate(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validate", reflect.TypeOf((*MockConsensus)(nil).Validate), arg0)
}

// MockNetService is a mock of NetService interface
type MockNetService struct {
	ctrl     *gomock.Controller
	recorder *MockNetServiceMockRecorder
}

// MockNetServiceMockRecorder is the mock recorder for MockNetService
type MockNetServiceMockRecorder struct {
	mock *MockNetService
}

// NewMockNetService creates a new mock instance
func NewMockNetService(ctrl *gomock.Controller) *MockNetService {
	mock := &MockNetService{ctrl: ctrl}
	mock.recorder = &MockNetServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockNetService) EXPECT() *MockNetServiceMockRecorder {
	return m.recorder
}

// BroadcastBlock mocks base method
func (m *MockNetService) BroadcastBlock(arg0 *core.Block) error {
	ret := m.ctrl.Call(m, "BroadcastBlock", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// BroadcastBlock indicates an expected call of BroadcastBlock
func (mr *MockNetServiceMockRecorder) BroadcastBlock(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BroadcastBlock", reflect.TypeOf((*MockNetService)(nil).BroadcastBlock), arg0)
}

// GetBlockchain mocks base method
func (m *MockNetService) GetBlockchain() *core.Blockchain {
	ret := m.ctrl.Call(m, "GetBlockchain")
	ret0, _ := ret[0].(*core.Blockchain)
	return ret0
}

// GetBlockchain indicates an expected call of GetBlockchain
func (mr *MockNetServiceMockRecorder) GetBlockchain() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockchain", reflect.TypeOf((*MockNetService)(nil).GetBlockchain))
}

// GetPeerID mocks base method
func (m *MockNetService) GetPeerID() go_libp2p_peer.ID {
	ret := m.ctrl.Call(m, "GetPeerID")
	ret0, _ := ret[0].(go_libp2p_peer.ID)
	return ret0
}

// GetPeerID indicates an expected call of GetPeerID
func (mr *MockNetServiceMockRecorder) GetPeerID() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPeerID", reflect.TypeOf((*MockNetService)(nil).GetPeerID))
}

// MockBlockPoolInterface is a mock of BlockPoolInterface interface
type MockBlockPoolInterface struct {
	ctrl     *gomock.Controller
	recorder *MockBlockPoolInterfaceMockRecorder
}

// MockBlockPoolInterfaceMockRecorder is the mock recorder for MockBlockPoolInterface
type MockBlockPoolInterfaceMockRecorder struct {
	mock *MockBlockPoolInterface
}

// NewMockBlockPoolInterface creates a new mock instance
func NewMockBlockPoolInterface(ctrl *gomock.Controller) *MockBlockPoolInterface {
	mock := &MockBlockPoolInterface{ctrl: ctrl}
	mock.recorder = &MockBlockPoolInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBlockPoolInterface) EXPECT() *MockBlockPoolInterfaceMockRecorder {
	return m.recorder
}

// BlockRequestCh mocks base method
func (m *MockBlockPoolInterface) BlockRequestCh() chan core.BlockRequestPars {
	ret := m.ctrl.Call(m, "BlockRequestCh")
	ret0, _ := ret[0].(chan core.BlockRequestPars)
	return ret0
}

// BlockRequestCh indicates an expected call of BlockRequestCh
func (mr *MockBlockPoolInterfaceMockRecorder) BlockRequestCh() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BlockRequestCh", reflect.TypeOf((*MockBlockPoolInterface)(nil).BlockRequestCh))
}

// GetBlockchain mocks base method
func (m *MockBlockPoolInterface) GetBlockchain() *core.Blockchain {
	ret := m.ctrl.Call(m, "GetBlockchain")
	ret0, _ := ret[0].(*core.Blockchain)
	return ret0
}

// GetBlockchain indicates an expected call of GetBlockchain
func (mr *MockBlockPoolInterfaceMockRecorder) GetBlockchain() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockchain", reflect.TypeOf((*MockBlockPoolInterface)(nil).GetBlockchain))
}

// GetSyncState mocks base method
func (m *MockBlockPoolInterface) GetSyncState() bool {
	ret := m.ctrl.Call(m, "GetSyncState")
	ret0, _ := ret[0].(bool)
	return ret0
}

// GetSyncState indicates an expected call of GetSyncState
func (mr *MockBlockPoolInterfaceMockRecorder) GetSyncState() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSyncState", reflect.TypeOf((*MockBlockPoolInterface)(nil).GetSyncState))
}

// Push mocks base method
func (m *MockBlockPoolInterface) Push(arg0 *core.Block, arg1 go_libp2p_peer.ID) {
	m.ctrl.Call(m, "Push", arg0, arg1)
}

// Push indicates an expected call of Push
func (mr *MockBlockPoolInterfaceMockRecorder) Push(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Push", reflect.TypeOf((*MockBlockPoolInterface)(nil).Push), arg0, arg1)
}

// SetBlockchain mocks base method
func (m *MockBlockPoolInterface) SetBlockchain(arg0 *core.Blockchain) {
	m.ctrl.Call(m, "SetBlockchain", arg0)
}

// SetBlockchain indicates an expected call of SetBlockchain
func (mr *MockBlockPoolInterfaceMockRecorder) SetBlockchain(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetBlockchain", reflect.TypeOf((*MockBlockPoolInterface)(nil).SetBlockchain), arg0)
}

// SetSyncState mocks base method
func (m *MockBlockPoolInterface) SetSyncState(arg0 bool) {
	m.ctrl.Call(m, "SetSyncState", arg0)
}

// SetSyncState indicates an expected call of SetSyncState
func (mr *MockBlockPoolInterfaceMockRecorder) SetSyncState(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetSyncState", reflect.TypeOf((*MockBlockPoolInterface)(nil).SetSyncState), arg0)
}

// VerifyTransactions mocks base method
func (m *MockBlockPoolInterface) VerifyTransactions(arg0 core.UTXOIndex, arg1 []*core.Block) bool {
	ret := m.ctrl.Call(m, "VerifyTransactions", arg0, arg1)
	ret0, _ := ret[0].(bool)
	return ret0
}

// VerifyTransactions indicates an expected call of VerifyTransactions
func (mr *MockBlockPoolInterfaceMockRecorder) VerifyTransactions(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyTransactions", reflect.TypeOf((*MockBlockPoolInterface)(nil).VerifyTransactions), arg0, arg1)
}
