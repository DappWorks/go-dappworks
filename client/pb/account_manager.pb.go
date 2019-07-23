// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/dappley/go-dappley/client/pb/account_manager.proto

package accountpb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type AccountManager struct {
	Accounts             []*Account `protobuf:"bytes,1,rep,name=accounts,proto3" json:"accounts,omitempty"`
	PassPhrase           []byte     `protobuf:"bytes,2,opt,name=passPhrase,proto3" json:"passPhrase,omitempty"`
	Locked               bool       `protobuf:"varint,3,opt,name=locked,proto3" json:"locked,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *AccountManager) Reset()         { *m = AccountManager{} }
func (m *AccountManager) String() string { return proto.CompactTextString(m) }
func (*AccountManager) ProtoMessage()    {}
func (*AccountManager) Descriptor() ([]byte, []int) {
	return fileDescriptor_b20b9af9539eb66a, []int{0}
}

func (m *AccountManager) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AccountManager.Unmarshal(m, b)
}
func (m *AccountManager) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AccountManager.Marshal(b, m, deterministic)
}
func (m *AccountManager) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccountManager.Merge(m, src)
}
func (m *AccountManager) XXX_Size() int {
	return xxx_messageInfo_AccountManager.Size(m)
}
func (m *AccountManager) XXX_DiscardUnknown() {
	xxx_messageInfo_AccountManager.DiscardUnknown(m)
}

var xxx_messageInfo_AccountManager proto.InternalMessageInfo

func (m *AccountManager) GetAccounts() []*Account {
	if m != nil {
		return m.Accounts
	}
	return nil
}

func (m *AccountManager) GetPassPhrase() []byte {
	if m != nil {
		return m.PassPhrase
	}
	return nil
}

func (m *AccountManager) GetLocked() bool {
	if m != nil {
		return m.Locked
	}
	return false
}

func init() {
	proto.RegisterType((*AccountManager)(nil), "accountpb.AccountManager")
}

func init() {
	proto.RegisterFile("github.com/dappley/go-dappley/client/pb/account_manager.proto", fileDescriptor_b20b9af9539eb66a)
}

var fileDescriptor_b20b9af9539eb66a = []byte{
	// 174 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xb2, 0x4d, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0x49, 0x2c, 0x28, 0xc8, 0x49, 0xad, 0xd4, 0x4f,
	0xcf, 0xd7, 0x85, 0x31, 0x93, 0x73, 0x32, 0x53, 0xf3, 0x4a, 0xf4, 0x0b, 0x92, 0xf4, 0x13, 0x93,
	0x93, 0xf3, 0x4b, 0xf3, 0x4a, 0xe2, 0x73, 0x13, 0xf3, 0x12, 0xd3, 0x53, 0x8b, 0xf4, 0x0a, 0x8a,
	0xf2, 0x4b, 0xf2, 0x85, 0x38, 0xa1, 0xc2, 0x05, 0x49, 0x52, 0xa6, 0x24, 0x9a, 0x04, 0x31, 0x41,
	0xa9, 0x82, 0x8b, 0xcf, 0x11, 0x22, 0xe0, 0x0b, 0x31, 0x59, 0x48, 0x8f, 0x8b, 0x03, 0xaa, 0xa4,
	0x58, 0x82, 0x51, 0x81, 0x59, 0x83, 0xdb, 0x48, 0x48, 0x0f, 0x6e, 0x8d, 0x1e, 0x54, 0x71, 0x10,
	0x5c, 0x8d, 0x90, 0x1c, 0x17, 0x57, 0x41, 0x62, 0x71, 0x71, 0x40, 0x46, 0x51, 0x62, 0x71, 0xaa,
	0x04, 0x93, 0x02, 0xa3, 0x06, 0x4f, 0x10, 0x92, 0x88, 0x90, 0x18, 0x17, 0x5b, 0x4e, 0x7e, 0x72,
	0x76, 0x6a, 0x8a, 0x04, 0xb3, 0x02, 0xa3, 0x06, 0x47, 0x10, 0x94, 0x97, 0xc4, 0x06, 0x76, 0x80,
	0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0x83, 0xa7, 0x35, 0x79, 0x03, 0x01, 0x00, 0x00,
}