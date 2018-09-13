// Code generated by MockGen. DO NOT EDIT.
// Source: interfaces.go

// Package core_mock is a generated GoMock package.
package core_mock

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

// Validate mocks base method
func (m *MockConsensus) Validate(block *core.Block) bool {
	ret := m.ctrl.Call(m, "Validate", block)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Validate indicates an expected call of Validate
func (mr *MockConsensusMockRecorder) Validate(block interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validate", reflect.TypeOf((*MockConsensus)(nil).Validate), block)
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

// StartNewBlockMinting mocks base method
func (m *MockConsensus) StartNewBlockMinting() {
	m.ctrl.Call(m, "StartNewBlockMinting")
}

// StartNewBlockMinting indicates an expected call of StartNewBlockMinting
func (mr *MockConsensusMockRecorder) StartNewBlockMinting() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartNewBlockMinting", reflect.TypeOf((*MockConsensus)(nil).StartNewBlockMinting))
}

// Setup mocks base method
func (m *MockConsensus) Setup(arg0 core.NetService, arg1 string) {
	m.ctrl.Call(m, "Setup", arg0, arg1)
}

// Setup indicates an expected call of Setup
func (mr *MockConsensusMockRecorder) Setup(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Setup", reflect.TypeOf((*MockConsensus)(nil).Setup), arg0, arg1)
}

// SetTargetBit mocks base method
func (m *MockConsensus) SetTargetBit(arg0 int) {
	m.ctrl.Call(m, "SetTargetBit", arg0)
}

// SetTargetBit indicates an expected call of SetTargetBit
func (mr *MockConsensusMockRecorder) SetTargetBit(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTargetBit", reflect.TypeOf((*MockConsensus)(nil).SetTargetBit), arg0)
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
func (m *MockNetService) BroadcastBlock(block *core.Block) error {
	ret := m.ctrl.Call(m, "BroadcastBlock", block)
	ret0, _ := ret[0].(error)
	return ret0
}

// BroadcastBlock indicates an expected call of BroadcastBlock
func (mr *MockNetServiceMockRecorder) BroadcastBlock(block interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BroadcastBlock", reflect.TypeOf((*MockNetService)(nil).BroadcastBlock), block)
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

// SetBlockchain mocks base method
func (m *MockBlockPoolInterface) SetBlockchain(bc *core.Blockchain) {
	m.ctrl.Call(m, "SetBlockchain", bc)
}

// SetBlockchain indicates an expected call of SetBlockchain
func (mr *MockBlockPoolInterfaceMockRecorder) SetBlockchain(bc interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetBlockchain", reflect.TypeOf((*MockBlockPoolInterface)(nil).SetBlockchain), bc)
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

// GetForkPool mocks base method
func (m *MockBlockPoolInterface) GetForkPool() []*core.Block {
	ret := m.ctrl.Call(m, "GetForkPool")
	ret0, _ := ret[0].([]*core.Block)
	return ret0
}

// GetForkPool indicates an expected call of GetForkPool
func (mr *MockBlockPoolInterfaceMockRecorder) GetForkPool() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetForkPool", reflect.TypeOf((*MockBlockPoolInterface)(nil).GetForkPool))
}

// ForkPoolLen mocks base method
func (m *MockBlockPoolInterface) ForkPoolLen() int {
	ret := m.ctrl.Call(m, "ForkPoolLen")
	ret0, _ := ret[0].(int)
	return ret0
}

// ForkPoolLen indicates an expected call of ForkPoolLen
func (mr *MockBlockPoolInterfaceMockRecorder) ForkPoolLen() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ForkPoolLen", reflect.TypeOf((*MockBlockPoolInterface)(nil).ForkPoolLen))
}

// GetForkPoolHeadBlk mocks base method
func (m *MockBlockPoolInterface) GetForkPoolHeadBlk() *core.Block {
	ret := m.ctrl.Call(m, "GetForkPoolHeadBlk")
	ret0, _ := ret[0].(*core.Block)
	return ret0
}

// GetForkPoolHeadBlk indicates an expected call of GetForkPoolHeadBlk
func (mr *MockBlockPoolInterfaceMockRecorder) GetForkPoolHeadBlk() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetForkPoolHeadBlk", reflect.TypeOf((*MockBlockPoolInterface)(nil).GetForkPoolHeadBlk))
}

// GetForkPoolTailBlk mocks base method
func (m *MockBlockPoolInterface) GetForkPoolTailBlk() *core.Block {
	ret := m.ctrl.Call(m, "GetForkPoolTailBlk")
	ret0, _ := ret[0].(*core.Block)
	return ret0
}

// GetForkPoolTailBlk indicates an expected call of GetForkPoolTailBlk
func (mr *MockBlockPoolInterfaceMockRecorder) GetForkPoolTailBlk() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetForkPoolTailBlk", reflect.TypeOf((*MockBlockPoolInterface)(nil).GetForkPoolTailBlk))
}

// ResetForkPool mocks base method
func (m *MockBlockPoolInterface) ResetForkPool() {
	m.ctrl.Call(m, "ResetForkPool")
}

// ResetForkPool indicates an expected call of ResetForkPool
func (mr *MockBlockPoolInterfaceMockRecorder) ResetForkPool() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResetForkPool", reflect.TypeOf((*MockBlockPoolInterface)(nil).ResetForkPool))
}

// ReInitializeForkPool mocks base method
func (m *MockBlockPoolInterface) ReInitializeForkPool(blk *core.Block) {
	m.ctrl.Call(m, "ReInitializeForkPool", blk)
}

// ReInitializeForkPool indicates an expected call of ReInitializeForkPool
func (mr *MockBlockPoolInterfaceMockRecorder) ReInitializeForkPool(blk interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReInitializeForkPool", reflect.TypeOf((*MockBlockPoolInterface)(nil).ReInitializeForkPool), blk)
}

// IsParentOfFork mocks base method
func (m *MockBlockPoolInterface) IsParentOfFork(blk *core.Block) bool {
	ret := m.ctrl.Call(m, "IsParentOfFork", blk)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsParentOfFork indicates an expected call of IsParentOfFork
func (mr *MockBlockPoolInterfaceMockRecorder) IsParentOfFork(blk interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsParentOfFork", reflect.TypeOf((*MockBlockPoolInterface)(nil).IsParentOfFork), blk)
}

// IsTailOfFork mocks base method
func (m *MockBlockPoolInterface) IsTailOfFork(blk *core.Block) bool {
	ret := m.ctrl.Call(m, "IsTailOfFork", blk)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsTailOfFork indicates an expected call of IsTailOfFork
func (mr *MockBlockPoolInterfaceMockRecorder) IsTailOfFork(blk interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsTailOfFork", reflect.TypeOf((*MockBlockPoolInterface)(nil).IsTailOfFork), blk)
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

// VerifyTransactions mocks base method
func (m *MockBlockPoolInterface) VerifyTransactions(utxo core.UtxoIndex) bool {
	ret := m.ctrl.Call(m, "VerifyTransactions", utxo)
	ret0, _ := ret[0].(bool)
	return ret0
}

// VerifyTransactions indicates an expected call of VerifyTransactions
func (mr *MockBlockPoolInterfaceMockRecorder) VerifyTransactions(utxo interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyTransactions", reflect.TypeOf((*MockBlockPoolInterface)(nil).VerifyTransactions), utxo)
}

// IsHigherThanFork mocks base method
func (m *MockBlockPoolInterface) IsHigherThanFork(block *core.Block) bool {
	ret := m.ctrl.Call(m, "IsHigherThanFork", block)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsHigherThanFork indicates an expected call of IsHigherThanFork
func (mr *MockBlockPoolInterfaceMockRecorder) IsHigherThanFork(block interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsHigherThanFork", reflect.TypeOf((*MockBlockPoolInterface)(nil).IsHigherThanFork), block)
}

// Push mocks base method
func (m *MockBlockPoolInterface) Push(block *core.Block, pid go_libp2p_peer.ID) {
	m.ctrl.Call(m, "Push", block, pid)
}

// Push indicates an expected call of Push
func (mr *MockBlockPoolInterfaceMockRecorder) Push(block, pid interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Push", reflect.TypeOf((*MockBlockPoolInterface)(nil).Push), block, pid)
}
