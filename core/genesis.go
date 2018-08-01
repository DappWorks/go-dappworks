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

const genesisCoinbaseData = "Hello world"


func NewGenesisBlock(address string) *Block {
	//return consensus.ProduceBlock(address, genesisCoinbaseData,[]byte{})

	txin := TXInput{nil, -1, nil, []byte(genesisCoinbaseData)}
	txout := NewTXOutput(subsidy, address)
	txs := []*Transaction{}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}, 0}
	tx.ID = tx.Hash()
	txs = append(txs,&tx)

	header := &BlockHeader{
		hash: []byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0},
		prevHash: []byte{},
		nonce:     0,
		timestamp: 1532392928, //July 23,2018 17:42 PST
	}
	b := &Block{
		header: header,
		transactions: txs,
		height:0,
	}

	b.SetHash(b.CalculateHash())
	return b
}
