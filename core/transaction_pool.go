// Copyright (C) 2018 go-dappley authors
//
// This file is part of the go-dappley library.
//
// the go-dappley library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either pubKeyHash 3 of the License, or
// (at your option) any later pubKeyHash.
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
	"encoding/gob"
	"github.com/dappley/go-dappley/storage"
	"sort"
	"sync"

	"github.com/asaskevich/EventBus"
	"github.com/golang-collections/collections/stack"
	logger "github.com/sirupsen/logrus"
)

const (
	NewTransactionTopic   = "NewTransaction"
	EvictTransactionTopic = "EvictTransaction"
	scheduleFuncName      = "dapp_schedule"
	TxPoolDbKey           = "txpool"
)

type TransactionNode struct {
	Children map[string]*Transaction
	Value    *Transaction
}

type TransactionPool struct {
	txs      map[string]*TransactionNode
	tipOrder []string
	limit    uint32
	EventBus EventBus.Bus
	mutex    sync.RWMutex
}

func NewTransactionPool(limit uint32) *TransactionPool {
	return &TransactionPool{
		txs:      make(map[string]*TransactionNode),
		tipOrder: make([]string, 0),
		limit:    limit,
		EventBus: EventBus.New(),
		mutex:    sync.RWMutex{},
	}
}

func (txPool *TransactionPool) deepCopy() *TransactionPool {
	txPoolCopy := TransactionPool{
		txs:      make(map[string]*TransactionNode),
		tipOrder: make([]string, 0),
		limit:    txPool.limit,
		EventBus: EventBus.New(),
		mutex:    sync.RWMutex{},
	}

	copy(txPoolCopy.tipOrder, txPool.tipOrder)

	for key, tx := range txPool.txs {
		newTx := tx.Value.DeepCopy()
		newTxNode := TransactionNode{Children: make(map[string]*Transaction), Value: &newTx}

		for childKey, childTx := range tx.Children {
			newTxNode.Children[childKey] = childTx
		}
		txPoolCopy.txs[key] = &newTxNode
	}

	return &txPoolCopy
}

func (txPool *TransactionPool) GetTransactions() []*Transaction {
	txPool.mutex.RLock()
	defer txPool.mutex.RUnlock()

	return txPool.getSortedTransactions()
}

//PopTransactionsWithMostTips pops the transactions with the most tips
func (txPool *TransactionPool) PopTransactionsWithMostTips(utxoIndex *UTXOIndex, numOfTx int) []*Transaction {

	tempUtxoIndex := utxoIndex.DeepCopy()
	var validTxs []*Transaction
	var inValidTxs []*Transaction

	for len(validTxs)<numOfTx && len(txPool.txs)>0 {
		txNode := txPool.GetMaxTipTransaction()
		txPool.tipOrder = txPool.tipOrder[1:]
		if txNode.Value.Verify(tempUtxoIndex, 0) {
			validTxs = append(validTxs, txNode.Value)
			tempUtxoIndex.UpdateUtxo(txNode.Value)
			txPool.insertChildrenIntoSortedWaitlist(txNode)
			txPool.removeTransaction(txNode.Value)
		} else {
			inValidTxs = append(inValidTxs, txNode.Value)
			txPool.removeTransactionNodeAndChildren(txNode.Value)
		}
	}

	return validTxs
}

func (txPool *TransactionPool) Push(tx *Transaction) {
	txPool.mutex.Lock()
	defer txPool.mutex.Unlock()
	if txPool.limit == 0 {
		logger.Warn("TransactionPool: transaction is not pushed to pool because limit is set to 0.")
		return
	}

	if len(txPool.txs) >= int(txPool.limit) {
		logger.WithFields(logger.Fields{
			"limit": txPool.limit,
		}).Warn("TransactionPool: is full.")

		minTx := txPool.GetMinTipTransaction()
		if minTx != nil && tx.Tip.Cmp(minTx.Value.Tip) < 1 {
			return
		}

		toRemoveTxs := txPool.getDependentTxs(minTx)
		if checkDependTxInMap(tx, toRemoveTxs) == true {
			logger.Warn("TransactionPool: failed to push because dependent transactions are not removed from pool.")
			return
		}

		txPool.removeMinTipTx()
	}

	txPool.addTransaction(tx)

}

//CleanUpMinedTxs updates the transaction pool when a new block is added to the blockchain.
//It removes the packed transactions from the txpool while keeping their children
func (txPool *TransactionPool) CleanUpMinedTxs(minedTxs []*Transaction) {
	txPool.mutex.Lock()
	defer txPool.mutex.Unlock()

	for _, tx := range minedTxs {

		txNode, ok := txPool.txs[string(tx.ID)]
		if !ok {
			continue
		}
		txPool.insertChildrenIntoSortedWaitlist(txNode)
		txPool.removeTransaction(txNode.Value)
	}
	txPool.cleanUpTxSort()
}

func (txPool *TransactionPool) cleanUpTxSort() {
	newTxOrder := []string{}
	for _, txid := range txPool.tipOrder {
		if _, ok := txPool.txs[txid]; ok {
			newTxOrder = append(newTxOrder, txid)
		}
	}
	txPool.tipOrder = newTxOrder
}

func (txPool *TransactionPool) getSortedTransactions() []*Transaction {
	nodes := make(map[string]*TransactionNode)
	isExecTxOkToInsert := true

	for key, node := range txPool.txs {
		nodes[key] = node
		if node.Value.IsContract() && !node.Value.IsExecutionContract() {
			isExecTxOkToInsert = false
		}
	}

	var sortedTxs []*Transaction
	for len(nodes) > 0 {
		for key, node := range nodes {
			if !checkDependTxInMap(node.Value, nodes) {
				if node.Value.IsContract() {
					if node.Value.IsExecutionContract() {
						if isExecTxOkToInsert {
							sortedTxs = append(sortedTxs, node.Value)
							delete(nodes, key)
						}
					} else {
						sortedTxs = append(sortedTxs, node.Value)
						delete(nodes, key)
						isExecTxOkToInsert = true
					}
				} else {
					sortedTxs = append(sortedTxs, node.Value)
					delete(nodes, key)
				}
			}
		}
	}

	return sortedTxs
}

func checkDependTxInMap(tx *Transaction, existTxs map[string]*TransactionNode) bool {
	for _, vin := range tx.Vin {
		if _, exist := existTxs[string(vin.Txid)]; exist {
			return true
		}
	}
	return false
}

func (txPool *TransactionPool) GetTransactionById(txid []byte) *Transaction {
	txPool.mutex.RLock()
	txPool.mutex.RUnlock()
	return txPool.txs[string(txid)].Value
}

func (txPool *TransactionPool) getDependentTxs(txNode *TransactionNode) map[string]*TransactionNode {

	toRemoveTxs := make(map[string]*TransactionNode)
	toCheckTxs := []*TransactionNode{txNode}

	for len(toCheckTxs) > 0 {
		currentTxNode := toCheckTxs[0]
		toCheckTxs = toCheckTxs[1:]
		for key, _ := range currentTxNode.Children {
			toCheckTxs = append(toCheckTxs, txPool.txs[key])
		}
		toRemoveTxs[string(currentTxNode.Value.ID)] = currentTxNode
	}

	return toRemoveTxs
}

// The param toRemoveTxs must be calculate by function getDependentTxs
func (txPool *TransactionPool) removeSelectedTransactions(toRemoveTxs map[string]*TransactionNode) {
	for _, txNode := range toRemoveTxs {
		txPool.removeTransactionNodeAndChildren(txNode.Value)
	}
}

//removeTransactionNodeAndChildren removes the txNode from tx pool and all its children.
//Note: this function does not remove the node from tipOrder!
func (txPool *TransactionPool) removeTransactionNodeAndChildren(tx *Transaction) {

	txStack := stack.New()
	txStack.Push(string(tx.ID))
	for(txStack.Len()>0){
		txid := txStack.Pop().(string)
		currTxNode, ok := txPool.txs[txid]
		if !ok{
			continue
		}
		for _, child := range currTxNode.Children {
			txStack.Push(string(child.ID))
		}
		txPool.EventBus.Publish(EvictTransactionTopic, txPool.txs[txid].Value)
		txPool.removeTransaction(currTxNode.Value)
	}
}

//removeTransactionNodeAndChildren removes the txNode from tx pool.
//Note: this function does not remove the node from tipOrder!
func (txPool *TransactionPool) removeTransaction(tx *Transaction) {
	txPool.disconnectFromParent(tx)
	txPool.EventBus.Publish(EvictTransactionTopic, txPool.txs[string(tx.ID)].Value)
	delete(txPool.txs, string(tx.ID))
}

//disconnectFromParent removes itself from its parent's node's children field
func (txPool *TransactionPool) disconnectFromParent(tx *Transaction){
	for _, vin := range tx.Vin {
		if parentTx, exist := txPool.txs[string(vin.Txid)]; exist {
			delete(parentTx.Children, string(tx.ID))
		}
	}
}

func (txPool *TransactionPool) removeMinTipTx() {
	minTipTx := txPool.GetMinTipTransaction()
	if minTipTx == nil {
		return
	}
	txPool.removeTransactionNodeAndChildren(minTipTx.Value)
	txPool.tipOrder = txPool.tipOrder[:len(txPool.tipOrder)-1]
}

func (txPool *TransactionPool) addTransaction(tx *Transaction) {
	isDependentOnParent := false
	for _, vin := range tx.Vin {
		parentTx, exist := txPool.txs[string(vin.Txid)]
		if exist {
			parentTx.Children[string(tx.ID)] = tx
			isDependentOnParent = true
		}
	}

	txNode := TransactionNode{Children: make(map[string]*Transaction), Value: tx}
	txPool.txs[string(tx.ID)] = &txNode

	txPool.EventBus.Publish(NewTransactionTopic, tx)

	//if it depends on another tx in txpool, the transaction will be not be included in the sorted list
	if isDependentOnParent {
		return
	}

	txPool.insertIntoSortedWaitlist(tx)
}

func (txPool *TransactionPool) insertChildrenIntoSortedWaitlist(txNode *TransactionNode){
	for _, child := range txNode.Children {
		txPool.insertIntoSortedWaitlist(child)
	}
}

//insertIntoSortedWaitlist insert a transaction into txSort based on tip.
//If the transaction is a child of another transaction, the transaction will NOT be inserted
func (txPool *TransactionPool) insertIntoSortedWaitlist(tx *Transaction){
	index := sort.Search(len(txPool.tipOrder), func(i int) bool {
		return txPool.txs[txPool.tipOrder[i]].Value.Tip.Cmp(tx.Tip) == -1
	})

	txPool.tipOrder = append(txPool.tipOrder, "")
	copy(txPool.tipOrder[index+1:], txPool.tipOrder[index:])
	txPool.tipOrder[index] = string(tx.ID)
}

func deserializeTxPool(d []byte) map[string]*TransactionNode {
	txs := make(map[string]*TransactionNode)
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&txs)
	if err != nil {
		logger.WithError(err).Panic("TxPool: failed to deserialize TxPool transactions.")
	}
	return txs
}

func (txPool *TransactionPool) LoadFromDatabase(db storage.Storage) {
	txPool.mutex.Lock()
	defer txPool.mutex.Unlock()
	rawBytes, err := db.Get([]byte(TxPoolDbKey))
	if err != nil && err.Error() == storage.ErrKeyInvalid.Error() || len(rawBytes) == 0 {
		return
	}
	txPool.txs = deserializeTxPool(rawBytes)
}

func (txPool *TransactionPool) serialize() []byte {

	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(txPool.txs)
	if err != nil {
		logger.WithError(err).Panic("TxPool: failed to serialize TxPool transactions.")
	}
	return encoded.Bytes()
}

func (txPool *TransactionPool) SaveToDatabase(db storage.Storage) error {
	txPool.mutex.Lock()
	defer txPool.mutex.Unlock()
	return db.Put([]byte(TxPoolDbKey), txPool.serialize())
}

//GetMinTipTransaction gets the transactionNode with minimum tip
func (txPool *TransactionPool) GetMaxTipTransaction() *TransactionNode {
	txid := txPool.GetMaxTipTxid()
	if txid == "" {
		return nil
	}
	return txPool.txs[txid]
}

//GetMinTipTransaction gets the transactionNode with minimum tip
func (txPool *TransactionPool) GetMinTipTransaction() *TransactionNode {
	txid := txPool.GetMinTipTxid()
	if txid == "" {
		return nil
	}
	return txPool.txs[txid]
}

//GetMinTipTxid gets the txid of the transaction with minimum tip
func (txPool *TransactionPool) GetMaxTipTxid() string {
	if len(txPool.tipOrder) == 0 {
		return ""
	}
	return txPool.tipOrder[0]
}

//GetMinTipTxid gets the txid of the transaction with minimum tip
func (txPool *TransactionPool) GetMinTipTxid() string {
	if len(txPool.tipOrder) == 0 {
		return ""
	}
	return txPool.tipOrder[len(txPool.tipOrder)-1]
}
