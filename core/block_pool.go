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

	"github.com/dappley/go-dappley/common"
	"github.com/hashicorp/golang-lru"
	"github.com/libp2p/go-libp2p-peer"
	logger "github.com/sirupsen/logrus"
)

const BlockCacheLRUCacheLimit = 1024
const ForkCacheLRUCacheLimit = 128

type BlockRequestPars struct {
	BlockHash Hash
	Pid       peer.ID
}

type RcvedBlock struct {
	Block *Block
	Pid   peer.ID
}

type BlockPool struct {
	blockRequestCh chan BlockRequestPars
	size           int
	blockchain     *Blockchain
	blkCache       *lru.Cache //cache of full blks
}

func NewBlockPool(size int) *BlockPool {
	pool := &BlockPool{
		size:           size,
		blockRequestCh: make(chan BlockRequestPars, size),
		blockchain:     nil,
	}
	pool.blkCache, _ = lru.New(BlockCacheLRUCacheLimit)

	return pool
}

func (pool *BlockPool) SetBlockchain(bc *Blockchain) {
	pool.blockchain = bc
}

func (pool *BlockPool) BlockRequestCh() chan BlockRequestPars {
	return pool.blockRequestCh
}

func (pool *BlockPool) GetBlockchain() *Blockchain {
	return pool.blockchain
}

//Verify all transactions in a fork
func (pool *BlockPool) VerifyTransactions(utxo UTXOIndex, forkBlks []*Block) bool {
	for i := len(forkBlks) - 1; i >= 0; i-- {
		logger.Info("Start Verify")
		if !forkBlks[i].VerifyTransactions(utxo) {
			return false
		}
		logger.Info("Verifyed a block. Height: ", forkBlks[i].GetHeight(), "Have ", i, "block left")
		utxoIndex := LoadUTXOIndex(pool.blockchain.GetDb())
		utxoIndex.BuildForkUtxoIndex(forkBlks[i], pool.blockchain.GetDb())
	}
	return true
}

func (pool *BlockPool) Push(block *Block, pid peer.ID) {
	logger.Debug("BlockPool: Has received a new block")

	if !block.VerifyHash() {
		logger.Warn("BlockPool: The received block cannot pass hash verification!")
		return
	}

	if !(pool.blockchain.GetConsensus().VerifyBlock(block)) {
		logger.Warn("BlockPool: The received block cannot pass signature verification!")
		return
	}
	//TODO: Verify double spending transactions in the same block

	logger.Debug("BlockPool: Block has been verified")
	pool.handleRecvdBlock(block, pid)
}

func (pool *BlockPool) handleRecvdBlock(blk *Block, sender peer.ID) {
	logger.Debug("BlockPool: Received a new block: ", hex.EncodeToString(blk.GetHash()), " From Sender: ", sender.String())

	if !pool.blockchain.consensus.Validate(blk) {
		logger.Debug("BlockPool: Block: ", hex.EncodeToString(blk.GetHash()), " did not pass consensus validation, discarding block")
		return
	}

	tree, _ := common.NewTree(blk.hashString(), blk)
	blkCache := pool.blkCache

	if blkCache.Contains(blk.hashString()) {
		logger.Debug("BlockPool: BlockPool blkCache already contains blk: ", blk.GetHash(), " returning")
		return
	} else {
		logger.Debug("BlockPool: Adding blk key to blockcache: ", hex.EncodeToString(blk.GetHash()))
		blkCache.Add(blk.hashString(), tree)
	}

	bcTailBlk, err := pool.GetBlockchain().GetTailBlock()
	if err != nil {
		logger.Panicf("Can't find find any block %v", err)
	}

	pool.updatePoolForkCache(tree)

	if bcTailBlk.IsParentBlock(blk) {
		var forkTailTree *common.Tree
		_, forkTailTree = tree.FindHeightestChild(forkTailTree, 0, 0)
		trees := forkTailTree.GetParentTreesRange(tree)
		forkBlks := getBlocksFromTrees(trees)
		pool.blockchain.MergeFork(forkBlks)
		tree.Delete()
	} else {
		pool.requestPrevBlock(tree, sender)
	}
}

func getBlocksFromTrees(trees []*common.Tree) []*Block {
	var blocks []*Block
	for _, tree := range trees {
		blocks = append(blocks, tree.GetValue().(*Block))
	}
	return blocks
}

func (pool *BlockPool) updatePoolForkCache(tree *common.Tree) {
	blkCache := pool.blkCache
	// try to link children
	for _, key := range blkCache.Keys() {
		if possibleChild, ok := blkCache.Get(key); ok {
			if hex.EncodeToString(possibleChild.(*common.Tree).GetValue().(*Block).GetPrevHash()) == tree.GetValue().(*Block).hashString() {
				logger.Debug("BlockPool: Block: ", tree.GetValue().(*Block).hashString(), " found child Block: ", possibleChild.(*common.Tree).GetValue().(*Block).hashString(), " in BlockPool blkCache, adding child")
				tree.AddChild(possibleChild.(*common.Tree))
			}

		}
	}
	//link parent
	if parent, ok := blkCache.Get(string(tree.GetValue().(*Block).GetPrevHash())); ok == true {
		logger.Debug("BlockPool: Block: ", tree.GetValue().(*Block).hashString(), " found parent: ", hex.EncodeToString(tree.GetValue().(*Block).GetPrevHash()), " in BlockPool blkCache, adding parent")
		tree.AddParent(parent.(*common.Tree))
		*tree = *parent.(*common.Tree)
	}

	logger.Debug("BlockPool: Block: ", tree.GetValue().(*Block).hashString(), " finished updating BlockPoolCache")
}

func (pool *BlockPool) requestPrevBlock(tree *common.Tree, sender peer.ID) {
	logger.Debug("BlockPool: Block: ", tree.GetValue().(*Block).hashString(), " parent not found, proceeding to download parent: ", hex.EncodeToString(tree.GetValue().(*Block).GetPrevHash()), " from ", sender)
	pool.blockRequestCh <- BlockRequestPars{tree.GetValue().(*Block).GetPrevHash(), sender}
}

func (pool *BlockPool) getBlkFromBlkCache(hashString string) *Block {
	if val, ok := pool.blkCache.Get(hashString); ok == true {
		return val.(*Block)
	}
	return nil

}
