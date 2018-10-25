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
	"errors"
	"time"

	"github.com/dappley/go-dappley/client"
	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/network"
	"github.com/dappley/go-dappley/storage"
	logger "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const unlockduration = 300 * time.Second

var minerPrivateKey string
var (
	ErrInvalidAmount        = errors.New("ERROR: Amount is invalid (must be > 0)")
	ErrInvalidAddress       = errors.New("ERROR: Address is invalid")
	ErrInvalidSenderAddress = errors.New("ERROR: Sender address is invalid")
	ErrInvalidRcverAddress  = errors.New("ERROR: Receiver address is invalid")
	ErrPasswordNotMatch     = errors.New("ERROR: Password not correct")
	ErrPathEmpty            = errors.New("ERROR: Path empty")
	ErrPasswordEmpty        = errors.New("ERROR: Password empty")
)

//create a blockchain
func CreateBlockchain(address core.Address, db storage.Storage, consensus core.Consensus) (*core.Blockchain, error) {
	if !address.ValidateAddress() {
		return nil, ErrInvalidAddress
	}

	bc := core.CreateBlockchain(address, db, consensus)

	return bc, nil
}

//create a wallet from path
func CreateWallet(path string, password string) (*client.Wallet, error) {
	if len(path) == 0 {
		return nil, ErrPathEmpty
	}

	if len(password) == 0 {
		return nil, ErrPasswordEmpty
	}

	fl := storage.NewFileLoader(path)
	wm := client.NewWalletManager(fl)
	passBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	wm.PassPhrase = passBytes
	wm.Locked = true
	err = wm.LoadFromFile()
	wallet := client.NewWallet()
	wm.AddWallet(wallet)
	wm.SaveWalletToFile()

	return wallet, err
}

//get wallet
func GetWallet() (*client.Wallet, error) {
	wm, err := GetWalletManager(client.GetWalletFilePath())
	empty, err := wm.IsFileEmpty()
	if empty {
		return nil, nil
	}
	if len(wm.Wallets) > 0 {
		return wm.Wallets[0], err
	} else {
		return nil, err
	}
}

//Get lock flag
func IsWalletLocked() (bool, error) {
	wm, err := GetWalletManager(client.GetWalletFilePath())
	return wm.Locked, err
}

//Tell if the file empty or not exist
func IsWalletEmpty() (bool, error) {
	wm, _ := GetWalletManager(client.GetWalletFilePath())
	return wm.IsFileEmpty()
}

//Set lock flag
func SetLockWallet() error {
	wm, err1 := GetWalletManager(client.GetWalletFilePath())
	empty, err2 := wm.IsFileEmpty()
	if empty {
		return nil
	}
	if err1 != nil {
		return err1
	}

	if err2 != nil {
		return err2
	} else {
		wm.Locked = true
		wm.SaveWalletToFile()
		return nil
	}
}

//Set unlock and timer
func SetUnLockWallet() error {
	wm, err := GetWalletManager(client.GetWalletFilePath())
	if err != nil {
		return err
	} else {
		wm.SetUnlockTimer(unlockduration)
		return nil
	}
}

//create a wallet with passphrase
func CreateWalletWithpassphrase(password string) (*client.Wallet, error) {
	wm, err := GetWalletManager(client.GetWalletFilePath())
	if err != nil {
		return nil, err
	}

	if len(wm.Wallets) > 0 && wm.PassPhrase != nil {
		err = bcrypt.CompareHashAndPassword(wm.PassPhrase, []byte(password))
		if err != nil {
			return nil, ErrPasswordNotMatch
		}
		wallet := client.NewWallet()
		wm.AddWallet(wallet)
		wm.SaveWalletToFile()
		return wallet, err

	} else {
		passBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		wm.PassPhrase = passBytes
		logger.Info("Wallet password set!")
		wallet := client.NewWallet()
		wm.AddWallet(wallet)
		wm.Locked = true
		wm.SaveWalletToFile()
		return wallet, err
	}
}

//create a wallet
func AddWallet() (*client.Wallet, error) {
	wm, err := GetWalletManager(client.GetWalletFilePath())
	if err != nil {
		return nil, err
	}

	wallet := client.NewWallet()
	if len(wm.Wallets) == 0 {
		wm.Locked = true
	}
	wm.AddWallet(wallet)
	wm.SaveWalletToFile()
	return wallet, err
}

//Get duration
func GetUnlockDuration() time.Duration {
	return unlockduration
}

//get balance
func GetBalance(address core.Address, db storage.Storage) (*common.Amount, error) {
	pubKeyHash, valid := address.GetPubKeyHash()
	if valid == false {
		return common.NewAmount(0), ErrInvalidAddress
	}

	balance := common.NewAmount(0)
	utxoIndex := core.LoadUTXOIndex(db)
	utxos := utxoIndex.GetUTXOsByPubKeyHash(pubKeyHash)
	for _, out := range utxos {
		balance = balance.Add(out.Value)
	}

	return balance, nil
}

func Send(senderWallet *client.Wallet, to core.Address, amount *common.Amount, tip uint64, bc *core.Blockchain, node *network.Node) error {
	if !senderWallet.GetAddress().ValidateAddress() {
		return ErrInvalidSenderAddress
	}
	if !to.ValidateAddress() {
		return ErrInvalidRcverAddress
	}
	if amount.Validate() != nil || amount.IsZero() {
		return ErrInvalidAmount
	}

	tx, err := core.NewUTXOTransaction(bc.GetDb(), senderWallet.GetAddress(), to, amount, *senderWallet.GetKeyPair(), bc, tip)
	bc.GetTxPool().Push(tx)
	node.TxBroadcast(&tx)
	if err != nil {
		return err
	}

	return err
}

func SetMinerKeyPair(key string) {
	minerPrivateKey = key
}

func GetMinerAddress() string {
	return minerPrivateKey
}

//add balance
func AddBalance(address core.Address, amount *common.Amount, bc *core.Blockchain) error {
	if !address.ValidateAddress() {
		return ErrInvalidAddress
	}

	if amount.Validate() != nil || amount.IsZero() {
		return ErrInvalidAmount
	}
	minerKeyPair := core.GetKeyPairByString(minerPrivateKey)
	minerWallet := &client.Wallet{}
	minerWallet.Key = minerKeyPair
	minerWallet.Addresses = append(minerWallet.Addresses, minerWallet.Key.GenerateAddress())

	tx, err := core.NewUTXOTransaction(bc.GetDb(), minerWallet.GetAddress(), address, amount, *minerWallet.GetKeyPair(), bc, 0)

	if err != nil {
		return err
	}

	bc.GetTxPool().Push(tx)

	return err

}

func GetWalletManager(path string) (*client.WalletManager, error) {
	fl := storage.NewFileLoader(path)
	wm := client.NewWalletManager(fl)
	err := wm.LoadFromFile()
	if err != nil {
		return nil, err
	}
	return wm, nil
}
