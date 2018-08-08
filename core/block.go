// Copyright (C) 2018 go-dappley authors
//
// This file is part of the go-dappley library.
//
// the go-dappley library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-dappley library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU Gc
// eneral Public License
// along with the go-dappley library.  If not, see <http://www.gnu.org/licenses/>.
//
package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"

	"math/big"
	"reflect"

	"github.com/dappley/go-dappley/core/pb"
	"github.com/dappley/go-dappley/util"
	"github.com/gogo/protobuf/proto"
)

const (
	BlockVersion = 1
)

type BlockHeader struct {
	version int32
	hash      Hash
	prevHash  Hash
	nonce     int64
	timestamp int64
}

type Block struct {
	header       *BlockHeader
	transactions []*Transaction
	height       uint64
	checkFlag bool
}


func NewBlock(transactions []*Transaction, parent *Block) *Block {

	var prevHash []byte
	var height uint64
	var versionValue int32     //need set the version
	versionValue = BlockVersion
	height = 0
	if parent != nil {
		prevHash = parent.GetHash()
		height = parent.GetHeight() + 1
	}

	if transactions == nil {
		transactions = []*Transaction{}
	}
	return &Block{
		header: &BlockHeader{
			version: versionValue,
			hash:      []byte{},
			prevHash:  prevHash,
			nonce:     0,
			timestamp: time.Now().Unix(),
		},
		height:       height,
		transactions: transactions,
	}
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.transactions {
		txHashes = append(txHashes, tx.Hash())
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	bs := &BlockStream{
		Header: &BlockHeaderStream{
			Version: b.header.version,
			Hash:      b.header.hash,
			PrevHash:  b.header.prevHash,
			Nonce:     b.header.nonce,
			Timestamp: b.header.timestamp,
		},
		Transactions: b.transactions,
		Height:       b.height,
	}

	err := encoder.Encode(bs)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

func Deserialize(d []byte) *Block {
	var bs BlockStream
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&bs)
	if err != nil {
		log.Panic(err)
	}
	if bs.Header.Hash == nil {
		bs.Header.Hash = Hash{}
	}
	if bs.Header.PrevHash == nil {
		bs.Header.PrevHash = Hash{}
	}
	if bs.Header.MerkleRoot == nil {
		bs.Header.MerkleRoot = Hash{}
	}
	if bs.Transactions == nil {
		bs.Transactions = []*Transaction{}
	}
	return &Block{
		header: &BlockHeader{
			version:	bs.Header.Version,
			hash:      bs.Header.Hash,
			prevHash:  bs.Header.PrevHash,
			nonce:     bs.Header.Nonce,
			timestamp: bs.Header.Timestamp,
		},
		transactions: bs.Transactions,
		height:       bs.Height,
	}
}

func (b *Block) SetHash(hash Hash) {
	b.header.hash = hash
}

func (b *Block) GetHash() Hash {
	return b.header.hash
}

func (b *Block) GetHeight() uint64 {
	return b.height
}

func (b *Block) SetHeight(height uint64) {
	b.height = height
}

func (b *Block) GetPrevHash() Hash {
	return b.header.prevHash
}

func (b *Block) SetNonce(nonce int64) {
	b.header.nonce = nonce
}

func (b *Block) GetNonce() int64 {
	return b.header.nonce
}

func (b *Block) SetVersion(version int32) {
	b.header.version = version
}

func (b *Block) GetVersion() int32 {
	return b.header.version
}

func (b *Block) GetTimestamp() int64 {
	return b.header.timestamp
}

func (b *Block) GetTransactions() []*Transaction {
	return b.transactions
}

func (b *Block) ToProto() proto.Message {

	var txArray []*corepb.Transaction
	for _, tx := range b.transactions {
		txArray = append(txArray, tx.ToProto().(*corepb.Transaction))
	}

	return &corepb.Block{
		Header:       b.header.ToProto().(*corepb.BlockHeader),
		Transactions: txArray,
		Height:       b.height,
	}
}

func (b *Block) FromProto(pb proto.Message) {

	bh := BlockHeader{}
	bh.FromProto(pb.(*corepb.Block).Header)
	b.header = &bh

	var txs []*Transaction

	for _, txpb := range pb.(*corepb.Block).Transactions {
		tx := &Transaction{}
		tx.FromProto(txpb)
		txs = append(txs, tx)
	}
	b.transactions = txs

	b.height = pb.(*corepb.Block).Height
}

func (bh *BlockHeader) ToProto() proto.Message {
	return &corepb.BlockHeader{
		Version: bh.version,
		Hash:      bh.hash,
		Prevhash:  bh.prevHash,
		Nonce:     bh.nonce,
		Timestamp: bh.timestamp,
	}
}

func (bh *BlockHeader) FromProto(pb proto.Message) {
	bh.version = pb.(*corepb.BlockHeader).Version
	bh.hash = pb.(*corepb.BlockHeader).Hash
	bh.prevHash = pb.(*corepb.BlockHeader).Prevhash
	bh.nonce = pb.(*corepb.BlockHeader).Nonce
	bh.timestamp = pb.(*corepb.BlockHeader).Timestamp
}

func (b *Block) CalculateHash() Hash {
	return b.CalculateHashWithNonce(b.GetNonce())
}

func (b *Block) CalculateHashWithNonce(nonce int64) Hash {
	var hashInt big.Int

	data := bytes.Join(
		[][]byte{
			b.GetPrevHash(),
			b.HashTransactions(),
			util.IntToHex(b.GetTimestamp()),
			//util.IntToHex(targetBits),
			util.IntToHex(nonce),
		},
		[]byte{},
	)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	return hashInt.Bytes()
}

func (b *Block) VerifyHash() bool {
	return reflect.DeepEqual(b.GetHash(), b.CalculateHash())
}

func (b *Block) VerifyTransactions(utxo utxoIndex) bool {
	for _, tx := range b.GetTransactions() {
		if !tx.Verify(utxo) {
			return false
		}
	}
	return true
}

func IsParentBlockHash(parentBlk, childBlk *Block) bool{
	if parentBlk == nil || childBlk == nil{
		return false
	}
	return reflect.DeepEqual(parentBlk.GetHash(), childBlk.GetPrevHash())
}

func IsParentBlockHeight(parentBlk, childBlk *Block) bool{
	if parentBlk == nil || childBlk == nil{
		return false
	}
	return parentBlk.GetHeight() == childBlk.GetHeight()-1
}

func IsParentBlock(parentBlk, childBlk *Block) bool{
	return IsParentBlockHash(parentBlk, childBlk) && IsParentBlockHeight(parentBlk, childBlk)
}

func (b *Block) Rollback(){
	if b!= nil {
		txnPool := GetTxnPoolInstance()
		for _,tx := range b.GetTransactions(){
			if !tx.IsCoinbase() {
				txnPool.Push(*tx)
			}
		}
	}
}

func (b *Block) FindTransactionById(txid []byte) *Transaction{
	for _, tx := range b.GetTransactions() {
		if bytes.Compare(tx.ID, txid) == 0 {
			return tx
		}

	}
	return nil
}

//remove the transactions in a blook from the transaction pool
func (b *Block) RemoveMinedTxFromTxPool(){
	txPool := GetTxnPoolInstance()
	txPool.Traverse(func(tx Transaction) bool{
		//if the transaction in the transaction pool is not found in the block, keep it in the pool
		//if the transaction in the transaction pool is found in the block. remove it from the pool
		return b.FindTransactionById(tx.ID) == nil
	})
}