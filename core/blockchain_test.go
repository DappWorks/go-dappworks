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
// You should have received a copy of the GNU General Public License
// along with the go-dappley library.  If not, see <http://www.gnu.org/licenses/>.
//

package core

import (
	"encoding/hex"
	"errors"
	"github.com/dappley/go-dappley/storage"
	"github.com/dappley/go-dappley/storage/mocks"
	logger "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	logger.SetLevel(logger.WarnLevel)
	retCode := m.Run()
	os.Exit(retCode)
}

func TestCreateBlockchain(t *testing.T) {
	//create a new block chain
	s := storage.NewRamStorage()
	addr := NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	bc:= CreateBlockchain(addr, s,nil)

	//find next block. This block should be the genesis block and its prev hash should be empty
	blk,err := bc.Next()
	assert.Nil(t, err)
	assert.Empty(t, blk.GetPrevHash())
}

func TestBlockchain_HigherThanBlockchainTestHigher(t *testing.T) {
	//create a new block chain
	s := storage.NewRamStorage()
	addr := NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	bc:= CreateBlockchain(addr, s,nil)
	blk := GenerateMockBlock()
	blk.header.height = 1
	assert.True(t,bc.IsHigherThanBlockchain(blk))
}

func TestBlockchain_HigherThanBlockchainTestLower(t *testing.T) {
	//create a new block chain
	s := storage.NewRamStorage()
	addr := NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	bc:= CreateBlockchain(addr, s,nil)

	blk := GenerateMockBlock()
	blk.header.height = 1
	bc.AddBlockToTail(blk)

	assert.False(t,bc.IsHigherThanBlockchain(blk))
}

func TestBlockchain_IsInBlockchain(t *testing.T) {
	//create a new block chain
	s := storage.NewRamStorage()
	addr := NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	bc:= CreateBlockchain(addr, s,nil)

	blk := GenerateMockBlock()
	blk.SetHash([]byte("hash1"))
	blk.header.height = 1
	bc.AddBlockToTail(blk)

	isFound := bc.IsInBlockchain([]byte("hash1"))
	assert.True(t,isFound)

	isFound = bc.IsInBlockchain([]byte("hash2"))
	assert.False(t,isFound)
}

func TestBlockchain_RollbackToABlock(t *testing.T) {
	//create a mock blockchain with max height of 5
	bc := GenerateMockBlockchain(5)
	defer bc.db.Close()

	blk,err := bc.GetTailBlock()
	assert.Nil(t,err)

	//find the hash at height 3 (5-2)
	for i:=0; i<2; i++{
		blk,err = bc.GetBlockByHash(blk.GetPrevHash())
		assert.Nil(t,err)
	}

	//rollback to height 3
	bc.Rollback(blk.GetHash())

	//the height 3 block should be the new tail block
	newTailBlk,err := bc.GetTailBlock()
	assert.Nil(t,err)
	assert.Equal(t,blk.GetHash(),newTailBlk.GetHash())

}

func TestBlockchain_AddBlockToTail(t *testing.T) {

	// Serialized data of an empty UTXOIndex (generated using `hex.EncodeToString(UTXOIndex{}.serialize())`)
	serializedUTXOIndex, _ := hex.DecodeString("1aff87040101095554584f496e64657801ff8800010c01ff8600000dff8502010" +
		"2ff860001ff8200003bff81030102ff82000104010556616c756501ff8400010a5075624b657948617368010a00010454786964010a" +
		"0001075478496e64657801040000000aff83050102ff8a0000000fff8b05010103496e7401ff8c00000004ff880000")

	db := new(mocks.Storage)

	// Storage will allow blockchain creation to succeed
	db.On("Put", mock.Anything, mock.Anything).Return(nil)
	db.On("Get", []byte("utxo")).Return(serializedUTXOIndex, nil)
	db.On("EnableBatch").Return()
	db.On("DisableBatch").Return()
	db.On("Flush").Return(nil).Once()

	// Create a blockchain for testing
	addr := NewAddress("16PencPNnF8CiSx2EBGEd1axhf7vuHCouj")
	bc := &Blockchain{Hash{}, db, nil, nil, nil, nil}

	// Add genesis block
	genesis := NewGenesisBlock(addr.Address)
	err := bc.AddBlockToTail(genesis)

	// Expect batch write was used
	db.AssertCalled(t, "EnableBatch")
	db.AssertCalled(t, "Flush")
	db.AssertCalled(t, "DisableBatch")

	// Expect no error when adding genesis block
	assert.Nil(t, err)
	// Expect that blockchain tail is genesis block
	assert.Equal(t, genesis.GetHash(), Hash(bc.tailBlockHash))

	// Simulate a failure when flushing new block to storage
	simulatedFailure := errors.New("simulated storage failure")
	db.On("Flush").Return(simulatedFailure)

	// Add new block
	blk := GenerateMockBlock()
	blk.SetHash([]byte("hash1"))
	blk.header.height = 1
	err = bc.AddBlockToTail(blk)

	// Expect the simulated error when adding new block
	assert.Equal(t, simulatedFailure, err)
	// Expect that genesis block is still the blockchain tail
	assert.Equal(t, genesis.GetHash(), Hash(bc.tailBlockHash))

}
