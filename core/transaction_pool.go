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
	"bytes"

	"github.com/dappley/go-dappley/common/sorted"
	logger "github.com/sirupsen/logrus"
)

const TransactionPoolLimit = 128

type TransactionPool struct {
	messageCh    chan string
	exitCh       chan bool
	size         int
	Transactions sorted.Slice
}

func NewTransactionPool() *TransactionPool {
	txPool := &TransactionPool{
		messageCh: make(chan string, 128),
		size:      128,
	}
	txPool.Transactions = *sorted.NewSlice(CompareTransactionTips, match)
	return txPool
}

func CompareTransactionTips(a interface{}, b interface{}) int {
	ai := a.(Transaction)
	bi := b.(Transaction)
	if ai.Tip < bi.Tip {
		return -1
	} else if ai.Tip > bi.Tip {
		return 1
	} else {
		return 0
	}
}

// match returns true if a and b are Transactions and they have the same ID, false otherwise
func match(a interface{}, b interface{}) bool {
	return bytes.Compare(a.(Transaction).ID, b.(Transaction).ID) == 0
}

func (txPool *TransactionPool) RemoveMultipleTransactions(txs []*Transaction) {
	for _, tx := range txs {
		txPool.Transactions.Del(*tx)
	}
}

//function f should return true if the transaction needs to be pushed back to the pool
func (txPool *TransactionPool) Traverse(txHandler func(tx Transaction) bool) {

	for _, v := range txPool.Transactions.Get() {
		tx := v.(Transaction)
		if !txHandler(tx) {
			txPool.Transactions.Del(tx)
		}
	}
}

func (txPool *TransactionPool) FilterAllTransactions(utxoPool UTXOIndex) {
	txPool.Traverse(func(tx Transaction) bool {
		return tx.Verify(utxoPool, 0) // all transactions in transaction pool have no blockHeight
		// TODO: also check if amount is valid
	})
}

//need to optimize
func (txPool *TransactionPool) PopSortedTransactions() []*Transaction {
	sortedTransactions := []*Transaction{}
	for txPool.Transactions.Len() > 0 {
		tx := txPool.Transactions.PopRight().(Transaction)
		sortedTransactions = append(sortedTransactions, &tx)
	}
	return sortedTransactions
}

func (txPool *TransactionPool) Push(tx Transaction) {
	//get smallest tip tx

	if txPool.Transactions.Len() >= TransactionPoolLimit {
		compareTx := txPool.Transactions.PopLeft().(Transaction)
		greaterThanLeastTip := tx.Tip > compareTx.Tip
		if greaterThanLeastTip {
			txPool.Transactions.Push(tx)
		} else { // do nothing, push back popped tx
			txPool.Transactions.Push(compareTx)
		}
	} else {
		txPool.Transactions.Push(tx)
	}
}

func (txPool *TransactionPool) Start() {
	go txPool.messageLoop()
}

func (txPool *TransactionPool) Stop() {
	txPool.exitCh <- true
}

//todo: will change the input from string to transaction
func (txPool *TransactionPool) PushTransaction(msg string) {
	//func (txPool *TransactionPool) PushTransaction(tx *Transaction){
	//	txPool.Push(tx)
	logger.Info(msg)
}

func (txPool *TransactionPool) messageLoop() {
	for {
		select {
		case <-txPool.exitCh:
			logger.Info("Quit Transaction Pool")
			return
		case msg := <-txPool.messageCh:
			txPool.PushTransaction(msg)
		}
	}
}
