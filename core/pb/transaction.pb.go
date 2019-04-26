// Code generated by protoc-gen-go. DO NOT EDIT.
// source: core/pb/transaction.proto

package corepb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Transactions struct {
	Transactions         []*Transaction `protobuf:"bytes,1,rep,name=transactions,proto3" json:"transactions,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Transactions) Reset()         { *m = Transactions{} }
func (m *Transactions) String() string { return proto.CompactTextString(m) }
func (*Transactions) ProtoMessage()    {}
func (*Transactions) Descriptor() ([]byte, []int) {
	return fileDescriptor_transaction_13356641c4ccdb89, []int{0}
}
func (m *Transactions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transactions.Unmarshal(m, b)
}
func (m *Transactions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transactions.Marshal(b, m, deterministic)
}
func (dst *Transactions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transactions.Merge(dst, src)
}
func (m *Transactions) XXX_Size() int {
	return xxx_messageInfo_Transactions.Size(m)
}
func (m *Transactions) XXX_DiscardUnknown() {
	xxx_messageInfo_Transactions.DiscardUnknown(m)
}

var xxx_messageInfo_Transactions proto.InternalMessageInfo

func (m *Transactions) GetTransactions() []*Transaction {
	if m != nil {
		return m.Transactions
	}
	return nil
}

type Transaction struct {
	Id                   []byte      `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Vin                  []*TXInput  `protobuf:"bytes,2,rep,name=vin,proto3" json:"vin,omitempty"`
	Vout                 []*TXOutput `protobuf:"bytes,3,rep,name=vout,proto3" json:"vout,omitempty"`
	Tip                  []byte      `protobuf:"bytes,4,opt,name=tip,proto3" json:"tip,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *Transaction) Reset()         { *m = Transaction{} }
func (m *Transaction) String() string { return proto.CompactTextString(m) }
func (*Transaction) ProtoMessage()    {}
func (*Transaction) Descriptor() ([]byte, []int) {
	return fileDescriptor_transaction_13356641c4ccdb89, []int{1}
}
func (m *Transaction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transaction.Unmarshal(m, b)
}
func (m *Transaction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transaction.Marshal(b, m, deterministic)
}
func (dst *Transaction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transaction.Merge(dst, src)
}
func (m *Transaction) XXX_Size() int {
	return xxx_messageInfo_Transaction.Size(m)
}
func (m *Transaction) XXX_DiscardUnknown() {
	xxx_messageInfo_Transaction.DiscardUnknown(m)
}

var xxx_messageInfo_Transaction proto.InternalMessageInfo

func (m *Transaction) GetId() []byte {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *Transaction) GetVin() []*TXInput {
	if m != nil {
		return m.Vin
	}
	return nil
}

func (m *Transaction) GetVout() []*TXOutput {
	if m != nil {
		return m.Vout
	}
	return nil
}

func (m *Transaction) GetTip() []byte {
	if m != nil {
		return m.Tip
	}
	return nil
}

type TXInput struct {
	Txid                 []byte   `protobuf:"bytes,1,opt,name=txid,proto3" json:"txid,omitempty"`
	Vout                 int32    `protobuf:"varint,2,opt,name=vout,proto3" json:"vout,omitempty"`
	Signature            []byte   `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
	PublicKey            []byte   `protobuf:"bytes,4,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TXInput) Reset()         { *m = TXInput{} }
func (m *TXInput) String() string { return proto.CompactTextString(m) }
func (*TXInput) ProtoMessage()    {}
func (*TXInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_transaction_13356641c4ccdb89, []int{2}
}
func (m *TXInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TXInput.Unmarshal(m, b)
}
func (m *TXInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TXInput.Marshal(b, m, deterministic)
}
func (dst *TXInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TXInput.Merge(dst, src)
}
func (m *TXInput) XXX_Size() int {
	return xxx_messageInfo_TXInput.Size(m)
}
func (m *TXInput) XXX_DiscardUnknown() {
	xxx_messageInfo_TXInput.DiscardUnknown(m)
}

var xxx_messageInfo_TXInput proto.InternalMessageInfo

func (m *TXInput) GetTxid() []byte {
	if m != nil {
		return m.Txid
	}
	return nil
}

func (m *TXInput) GetVout() int32 {
	if m != nil {
		return m.Vout
	}
	return 0
}

func (m *TXInput) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *TXInput) GetPublicKey() []byte {
	if m != nil {
		return m.PublicKey
	}
	return nil
}

type TXOutput struct {
	Value                []byte   `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	PublicKeyHash        []byte   `protobuf:"bytes,2,opt,name=public_key_hash,json=publicKeyHash,proto3" json:"public_key_hash,omitempty"`
	Contract             string   `protobuf:"bytes,3,opt,name=contract,proto3" json:"contract,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TXOutput) Reset()         { *m = TXOutput{} }
func (m *TXOutput) String() string { return proto.CompactTextString(m) }
func (*TXOutput) ProtoMessage()    {}
func (*TXOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_transaction_13356641c4ccdb89, []int{3}
}
func (m *TXOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TXOutput.Unmarshal(m, b)
}
func (m *TXOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TXOutput.Marshal(b, m, deterministic)
}
func (dst *TXOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TXOutput.Merge(dst, src)
}
func (m *TXOutput) XXX_Size() int {
	return xxx_messageInfo_TXOutput.Size(m)
}
func (m *TXOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_TXOutput.DiscardUnknown(m)
}

var xxx_messageInfo_TXOutput proto.InternalMessageInfo

func (m *TXOutput) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *TXOutput) GetPublicKeyHash() []byte {
	if m != nil {
		return m.PublicKeyHash
	}
	return nil
}

func (m *TXOutput) GetContract() string {
	if m != nil {
		return m.Contract
	}
	return ""
}

func init() {
	proto.RegisterType((*Transactions)(nil), "corepb.Transactions")
	proto.RegisterType((*Transaction)(nil), "corepb.Transaction")
	proto.RegisterType((*TXInput)(nil), "corepb.TXInput")
	proto.RegisterType((*TXOutput)(nil), "corepb.TXOutput")
}

func init() {
	proto.RegisterFile("core/pb/transaction.proto", fileDescriptor_transaction_13356641c4ccdb89)
}

var fileDescriptor_transaction_13356641c4ccdb89 = []byte{
	// 285 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x91, 0xcd, 0x4a, 0xf4, 0x30,
	0x14, 0x86, 0xe9, 0xcf, 0xcc, 0x37, 0x3d, 0xd3, 0xcf, 0x19, 0x8e, 0x2e, 0xa2, 0x28, 0xd4, 0x22,
	0xd2, 0x55, 0x07, 0x74, 0xe1, 0x25, 0xa8, 0xb8, 0x10, 0x8a, 0x0b, 0x77, 0x43, 0xfa, 0x83, 0x0d,
	0x0e, 0x49, 0x68, 0x92, 0x32, 0x73, 0xf7, 0xd2, 0xb4, 0xb6, 0x75, 0x97, 0x73, 0xde, 0x87, 0xe7,
	0x3d, 0x10, 0xb8, 0x2c, 0x44, 0x53, 0xed, 0x64, 0xbe, 0xd3, 0x0d, 0xe5, 0x8a, 0x16, 0x9a, 0x09,
	0x9e, 0xca, 0x46, 0x68, 0x81, 0xcb, 0x2e, 0x92, 0x79, 0xfc, 0x0c, 0xe1, 0xc7, 0x14, 0x2a, 0x7c,
	0x82, 0x70, 0x06, 0x2b, 0xe2, 0x44, 0x5e, 0xb2, 0x7e, 0x38, 0x4f, 0x7b, 0x3c, 0x9d, 0xb1, 0xd9,
	0x1f, 0x30, 0x3e, 0xc2, 0x7a, 0x16, 0xe2, 0x19, 0xb8, 0xac, 0x24, 0x4e, 0xe4, 0x24, 0x61, 0xe6,
	0xb2, 0x12, 0x6f, 0xc1, 0x6b, 0x19, 0x27, 0xae, 0xd5, 0x6d, 0x46, 0xdd, 0xe7, 0x2b, 0x97, 0x46,
	0x67, 0x5d, 0x86, 0x77, 0xe0, 0xb7, 0xc2, 0x68, 0xe2, 0x59, 0x66, 0x3b, 0x31, 0xef, 0x46, 0x77,
	0x90, 0x4d, 0x71, 0x0b, 0x9e, 0x66, 0x92, 0xf8, 0xd6, 0xdc, 0x3d, 0x63, 0x0e, 0xff, 0x06, 0x0f,
	0x22, 0xf8, 0xfa, 0x38, 0xf6, 0xda, 0x77, 0xb7, 0xb3, 0x5a, 0x37, 0x72, 0x92, 0xc5, 0x20, 0xb9,
	0x86, 0x40, 0xb1, 0x2f, 0x4e, 0xb5, 0x69, 0x2a, 0xe2, 0x59, 0x78, 0x5a, 0xe0, 0x0d, 0x80, 0x34,
	0xf9, 0x81, 0x15, 0xfb, 0xef, 0xea, 0x34, 0x34, 0x05, 0xfd, 0xe6, 0xad, 0x3a, 0xc5, 0x25, 0xac,
	0x7e, 0x6f, 0xc2, 0x0b, 0x58, 0xb4, 0xf4, 0x60, 0xaa, 0xa1, 0xb1, 0x1f, 0xf0, 0x1e, 0x36, 0x93,
	0x60, 0x5f, 0x53, 0x55, 0xdb, 0xf6, 0x30, 0xfb, 0x3f, 0x5a, 0x5e, 0xa8, 0xaa, 0xf1, 0x0a, 0x56,
	0x85, 0xe0, 0xba, 0xa1, 0x85, 0xb6, 0x57, 0x04, 0xd9, 0x38, 0xe7, 0x4b, 0xfb, 0x4f, 0x8f, 0x3f,
	0x01, 0x00, 0x00, 0xff, 0xff, 0xc8, 0x0e, 0x37, 0x76, 0xc4, 0x01, 0x00, 0x00,
}
