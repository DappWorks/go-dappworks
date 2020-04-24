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

package consensus

import (
	"github.com/dappley/go-dappley/logic/ltransaction"
	"testing"

	"github.com/dappley/go-dappley/core/transaction"

	"github.com/dappley/go-dappley/core/block"
	"github.com/dappley/go-dappley/logic/lblock"

	"github.com/stretchr/testify/assert"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/core/account"
)

func TestNewDpos(t *testing.T) {
	dpos := NewDPOS(nil)
	assert.Equal(t, 1, cap(dpos.stopCh))
}

func TestDpos_beneficiaryIsProducer(t *testing.T) {
	producers := []string{
		"121yKAXeG4cw6uaGCBYjWk9yTWmMkhcoDD",
		"1MeSBgufmzwpiJNLemUe1emxAussBnz7a7",
		"1LCn8D5W7DLV1CbKE3buuJgNJjSeoBw2ct"}

	cbtx := ltransaction.NewCoinbaseTX(account.NewAddress(producers[0]), "", 0, common.NewAmount(0))
	cbtxInvalidProducer := ltransaction.NewCoinbaseTX(account.NewAddress(producers[0]), "", 0, common.NewAmount(0))

	tests := []struct {
		name     string
		block    *block.Block
		expected bool
	}{
		{
			name: "BeneficiaryIsProducer",
			block: FakeNewBlockWithTimestamp(
				46,
				[]*transaction.Transaction{
					core.MockTransaction(),
					&cbtx,
				},
				nil,
			),
			expected: true,
		},
		{
			name: "ProducerNotAtItsTurn",
			block: FakeNewBlockWithTimestamp(
				44,
				[]*transaction.Transaction{
					core.MockTransaction(),
					&cbtx,
				},
				nil,
			),
			expected: false,
		},
		{
			name: "NotAProducer",
			block: FakeNewBlockWithTimestamp(
				44,
				[]*transaction.Transaction{
					core.MockTransaction(),
					&cbtxInvalidProducer,
				},
				nil,
			),
			expected: false,
		},
		{
			name:     "EmptyBlock",
			block:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dpos := NewDPOS(nil)
			dpos.SetDynasty(NewDynasty(producers, len(producers), defaultTimeBetweenBlk))
			assert.Equal(t, tt.expected, dpos.isProducerBeneficiary(tt.block))
		})
	}
}

func TestDPOS_isDoubleMint(t *testing.T) {
	dpos := NewDPOS(nil)
	dpos.SetDynasty(NewDynasty(nil, defaultMaxProducers, defaultTimeBetweenBlk))
	blk1Time := int64(1548979365)
	blk2Time := int64(1548979366)

	// Both timestamps fall in the same DPoS time slot
	assert.Equal(t, int(blk1Time/defaultTimeBetweenBlk), int(blk2Time/defaultTimeBetweenBlk))

	blk1 := FakeNewBlockWithTimestamp(blk1Time, []*transaction.Transaction{}, nil)
	dpos.cacheBlock(blk1)
	blk2 := FakeNewBlockWithTimestamp(blk2Time, []*transaction.Transaction{}, nil)

	assert.True(t, dpos.isDoubleMint(blk2))
}

func FakeNewBlockWithTimestamp(t int64, txs []*transaction.Transaction, parent *block.Block) *block.Block {
	var prevHash []byte
	var height uint64
	height = 0
	if parent != nil {
		prevHash = parent.GetHash()
		height = parent.GetHeight() + 1
	}

	if txs == nil {
		txs = []*transaction.Transaction{}
	}
	blk := block.NewBlockWithRawInfo(
		[]byte{},
		prevHash,
		0,
		t,
		height,
		txs)

	hash := lblock.CalculateHashWithNonce(blk)
	blk.SetHash(hash)
	return blk
}
