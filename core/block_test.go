package core

import (
	"testing"

	"github.com/dappley/go-dappley/core/pb"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestDeserialize(t *testing.T) {
	blockExpect1 := NewBlock([]*Transaction{&Transaction{}}, []byte{'a'})
	b1 := blockExpect1.Serialize()
	block1 := Deserialize(b1)
	assert.Equal(t, blockExpect1.transactions, block1.transactions)
	assert.Equal(t, blockExpect1.header, block1.header)

	blockExpect2 := NewBlock(nil, nil)
	b2 := blockExpect2.Serialize()
	block2 := Deserialize(b2)
	assert.Equal(t, blockExpect2.transactions, block2.transactions)
	assert.Equal(t, blockExpect2.header, block2.header)

	blockExpect3 := NewBlock([]*Transaction{}, []byte{})
	b3 := blockExpect3.Serialize()
	block3 := Deserialize(b3)
	assert.Equal(t, blockExpect3.transactions, block3.transactions)
	assert.Equal(t, blockExpect3.header, block3.header)

}

func TestSerialize(t *testing.T) {
	block := NewBlock([]*Transaction{&Transaction{}}, []byte{'a'})
	b := block.Serialize()
	assert.NotNil(t, b)
}

func TestHashTransactions(t *testing.T) {
	block := NewBlock([]*Transaction{&Transaction{}}, []byte{'a'})
	hash := block.HashTransactions()
	assert.NotNil(t, hash)
}

func TestNewBlock(t *testing.T) {
	block1 := NewBlock(nil, nil)
	assert.NotNil(t, block1.header.prevHash)
	assert.NotNil(t, block1.transactions)

	block2 := NewBlock(nil, []byte{})
	assert.NotNil(t, block2.header.prevHash)
	assert.Equal(t, 0, len(block2.header.prevHash))
	assert.NotNil(t, block2.transactions)

	block3 := NewBlock(nil, []byte{'a'})
	assert.NotNil(t, block3.header.prevHash)
	assert.Equal(t, 1, len(block3.header.prevHash))
	assert.Equal(t, []byte{'a'}[0], block3.header.prevHash[0])
	assert.NotNil(t, block3.transactions)

	block4 := NewBlock([]*Transaction{}, nil)
	assert.NotNil(t, block4.header.prevHash)
	assert.NotNil(t, block4.transactions)
	assert.Equal(t, 0, len(block4.transactions))

	block5 := NewBlock([]*Transaction{&Transaction{}}, nil)
	assert.NotNil(t, block5.header.prevHash)
	assert.Equal(t, 1, len(block5.transactions))
	assert.Equal(t, &Transaction{}, block5.transactions[0])
	assert.NotNil(t, block5.transactions)
}

func TestBlockHeader_Proto(t *testing.T) {
	bh1 := BlockHeader{
		[]byte("hash"),
		[]byte("hash"),
		1,
		2,
	}

	pb := bh1.ToProto()
	mpb, err := proto.Marshal(pb)
	assert.Nil(t, err)

	newpb := &corepb.BlockHeader{}
	err = proto.Unmarshal(mpb, newpb)
	assert.Nil(t, err)

	bh2 := BlockHeader{}
	bh2.FromProto(newpb)

	assert.Equal(t, bh1, bh2)
}

func TestBlock_Proto(t *testing.T) {

	b1 := GenerateMockBlock()

	pb := b1.ToProto()
	mpb, err := proto.Marshal(pb)
	assert.Nil(t, err)

	newpb := &corepb.Block{}
	err = proto.Unmarshal(mpb, newpb)
	assert.Nil(t, err)

	b2 := &Block{}
	b2.FromProto(newpb)

	assert.Equal(t, *b1, *b2)
}
