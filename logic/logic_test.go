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

package logic

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/dappley/go-dappley/client"
	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/storage"
	logger "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const InvalidAddress = "Invalid Address"

func TestMain(m *testing.M) {
	cleanUpDatabase()
	logger.SetLevel(logger.WarnLevel)
	retCode := m.Run()
	cleanUpDatabase()
	os.Exit(retCode)
}

func TestCreateWallet(t *testing.T) {
	wallet, err := CreateWallet(GetTestWalletPath(), "test")
	assert.Nil(t, err)
	pubKeyHash, ok := wallet.Addresses[0].GetPubKeyHash()
	assert.Equal(t, true, ok)
	walletPubKeyHash, err := core.NewUserPubKeyHash(wallet.Key.PublicKey)
	assert.Nil(t, err)
	assert.Equal(t, pubKeyHash, []byte(walletPubKeyHash))
}

func TestCreateWalletWithPassphrase(t *testing.T) {
	wallet, err := CreateWallet(GetTestWalletPath(), "test")
	assert.Nil(t, err)
	pubKeyHash, ok := wallet.Addresses[0].GetPubKeyHash()
	assert.Equal(t, true, ok)
	walletPubKeyHash, err := core.NewUserPubKeyHash(wallet.Key.PublicKey)
	assert.Nil(t, err)
	assert.Equal(t, pubKeyHash, []byte(walletPubKeyHash))
}

func TestCreateBlockchain(t *testing.T) {

	store := storage.NewRamStorage()
	defer store.Close()

	//create a wallet address
	addr := core.NewAddress("dGDrVKjCG3sdXtDUgWZ7Fp3Q97tLhqWivf")

	//create a blockchain
	_, err := CreateBlockchain(addr, store, nil, 128, nil)
	assert.Nil(t, err)
}

func TestLoopCreateBlockchain(t *testing.T) {

	store := storage.NewRamStorage()
	defer store.Close()

	//create a wallet address

	err := ErrInvalidAddress
	//create a blockchain loop
	for i := 0; i < 2000; i++ {
		err = nil
		wallet := client.NewWallet()
		wallet.Key = core.NewKeyPair()
		addr := wallet.Key.GenerateAddress(false)
		if !addr.ValidateAddress() {
			fmt.Println(i, addr)
			err = ErrInvalidAddress
			break
		}
	}
	assert.Nil(t, err)
}

//create a blockchain with invalid address
func TestCreateBlockchainWithInvalidAddress(t *testing.T) {
	store := storage.NewRamStorage()
	defer store.Close()

	//create a blockchain with an invalid address
	bc, err := CreateBlockchain(core.NewAddress(InvalidAddress), store, nil, 128, nil)
	assert.Equal(t, ErrInvalidAddress, err)
	assert.Nil(t, bc)
}

func TestGetBalance(t *testing.T) {
	store := storage.NewRamStorage()
	defer store.Close()

	//create a wallet address
	addr := core.NewAddress("dGDrVKjCG3sdXtDUgWZ7Fp3Q97tLhqWivf")
	//create a blockchain
	bc, err := CreateBlockchain(addr, store, nil, 128, nil)
	assert.Nil(t, err)
	assert.NotNil(t, bc)

	//The balance should be 10000000 after creating a blockchain
	balance, err := GetBalance(addr, store)
	assert.Nil(t, err)
	assert.Equal(t, common.NewAmount(10000000), balance)
}

func TestGetBalanceWithInvalidAddress(t *testing.T) {

	store := storage.NewRamStorage()
	defer store.Close()

	//create a wallet address
	addr := core.NewAddress("dGDrVKjCG3sdXtDUgWZ7Fp3Q97tLhqWivf")
	//create a blockchain
	bc, err := CreateBlockchain(addr, store, nil, 128, nil)
	assert.Nil(t, err)
	assert.NotNil(t, bc)

	//The balance should be 10000000 after creating a blockchain
	balance1, err := GetBalance(core.NewAddress("dG6HhzSdA5m7KqvJNszVSf8i5f4neAteSs"), store)
	assert.Nil(t, err)
	assert.Equal(t, common.NewAmount(0), balance1)

	balance2, err := GetBalance(core.NewAddress("dG6HhzSdA5m7KqvJNszVSf8i5f4neAtfSs"), store)
	assert.Equal(t, ErrInvalidAddress, err)
	assert.Equal(t, common.NewAmount(0), balance2)
}

func TestGetAllAddresses(t *testing.T) {
	cleanUpDatabase()

	store := storage.NewRamStorage()
	defer store.Close()

	expectedRes := []core.Address{}
	//create a wallet address
	wallet, err := CreateWallet(GetTestWalletPath(), "test")
	assert.NotEmpty(t, wallet)
	addr := wallet.GetAddress()
	expectedRes = append(expectedRes, addr)

	//create a blockchain
	bc, err := CreateBlockchain(addr, store, nil, 128, nil)
	assert.Nil(t, err)
	assert.NotNil(t, bc)

	//create 10 more addresses
	for i := 0; i < 2; i++ {
		//create a wallet address
		wallet, err = CreateWallet(GetTestWalletPath(), "test")
		addr = wallet.GetAddress()
		assert.NotEmpty(t, addr)
		assert.Nil(t, err)
		expectedRes = append(expectedRes, addr)
	}

	//get all addresses
	addrs, err := GetAllAddressesByPath(GetTestWalletPath())
	assert.Nil(t, err)

	//the length should be equal
	assert.Equal(t, len(expectedRes), len(addrs))
	assert.ElementsMatch(t, expectedRes, addrs)
	cleanUpDatabase()
}

func TestIsWalletEmptyWallet(t *testing.T) {
	cleanUpDatabase()
	wallet1, err := CreateWallet(GetTestWalletPath(), "test")
	assert.NotEmpty(t, wallet1)
	assert.Nil(t, err)
	empty, err := IsTestWalletEmpty()
	assert.Nil(t, err)
	assert.Equal(t, false, empty)
	cleanUpDatabase()
	empty, err = IsTestWalletEmpty()
	assert.Nil(t, err)
	assert.Equal(t, true, empty)

	//teardown :clean up database amd files
	cleanUpDatabase()
}

func TestDeleteInvalidWallet(t *testing.T) {
	//setup: clean up database and files
	cleanUpDatabase()
	//create wallets address
	wallet1, err := CreateWallet(GetTestWalletPath(), "test")
	assert.NotEmpty(t, wallet1)
	addr1 := wallet1.GetAddress()

	addressList := []core.Address{addr1}

	list, err := GetAllAddressesByPath(GetTestWalletPath())
	assert.Nil(t, err)
	assert.ElementsMatch(t, list, addressList)

	//teardown :clean up database amd files
	cleanUpDatabase()
}

func isSameBlockChain(bc1, bc2 *core.Blockchain) bool {
	if bc1 == nil || bc2 == nil {
		return false
	}

	bci1 := bc1.Iterator()
	bci2 := bc2.Iterator()
	if bc1.GetMaxHeight() != bc2.GetMaxHeight() {
		return false
	}

loop:
	for {
		blk1, _ := bci1.Next()
		blk2, _ := bci2.Next()
		if blk1 == nil || blk2 == nil {
			break loop
		}
		if !reflect.DeepEqual(blk1.GetHash(), blk2.GetHash()) {
			return false
		}
	}
	return true
}

func cleanUpDatabase() {
	client.RemoveWalletFile()
}
