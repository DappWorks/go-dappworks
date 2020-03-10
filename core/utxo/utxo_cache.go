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

package utxo

import (
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/storage"
	lru "github.com/hashicorp/golang-lru"
)

const UtxoCacheLRUCacheLimit = 1024

// UTXOCache holds temporary UTXOTx data
type UTXOCache struct {
	// key: address, value: UTXOTx
	contractCache *lru.Cache
	cache         *lru.Cache
	db            storage.Storage
}

func NewUTXOCache(db storage.Storage) *UTXOCache {
	utxoCache := &UTXOCache{
		contractCache: nil,
		cache:         nil,
		db:            db,
	}
	utxoCache.cache, _ = lru.New(UtxoCacheLRUCacheLimit)
	utxoCache.contractCache, _ = lru.New(UtxoCacheLRUCacheLimit)
	return utxoCache
}

// Return value from cache
func (utxoCache *UTXOCache) Get(pubKeyHash account.PubKeyHash) *UTXOTx {
	mapData, ok := utxoCache.cache.Get(string(pubKeyHash))
	if ok {
		return mapData.(*UTXOTx)
	}

	rawBytes, err := utxoCache.db.Get(pubKeyHash)

	var utxoTx UTXOTx
	if err == nil {
		utxoTx = DeserializeUTXOTx(rawBytes)
		utxoCache.cache.Add(string(pubKeyHash), &utxoTx)
	} else {
		utxoTx = NewUTXOTx()
	}

	for _, u := range utxoTx.Indices {
		if u.UtxoType == UtxoCreateContract {
			utxoCache.contractCache.Add(string(pubKeyHash), u)
		}
	}
	return &utxoTx
}

// Return value from cache
func (utxoCache *UTXOCache) GetContractCreateUtxo(pubKeyHash account.PubKeyHash) *UTXO {
	mapData, ok := utxoCache.contractCache.Get(string(pubKeyHash))
	if !ok {
		utxotx := utxoCache.Get(pubKeyHash)
		if len(utxotx.Indices) > 0 {
			mapData, ok = utxoCache.contractCache.Get(string(pubKeyHash))
		} else {
			return nil
		}
	}

	return mapData.(*UTXO)
}

// Add new data into cache
func (utxoCache *UTXOCache) Put(pubKeyHash account.PubKeyHash, value *UTXOTx) error {
	if pubKeyHash == nil {
		return account.ErrEmptyPublicKeyHash
	}
	err := utxoCache.putUtxoTx(pubKeyHash, value)
	if err != nil {
		return err
	}

	for _, u := range value.Indices {
		if u.UtxoType == UtxoCreateContract {
			utxoCache.contractCache.Add(string(pubKeyHash), u)
		}
	}
	return nil
}

// Add new data into cache
func (utxoCache *UTXOCache) putUtxoTx(pubKeyHash account.PubKeyHash, value *UTXOTx) error {
	savedUtxoTx := value.DeepCopy()
	mapData, ok := utxoCache.cache.Get(string(pubKeyHash))
	utxoCache.cache.Add(string(pubKeyHash), savedUtxoTx)
	err := utxoCache.db.Put(pubKeyHash, value.Serialize())
	if err == nil && ok {
		Free(mapData.(*UTXOTx))
	}
	return err
}

func (utxoCache *UTXOCache) Delete(pubKeyHash account.PubKeyHash) error {
	if pubKeyHash == nil {
		return account.ErrEmptyPublicKeyHash
	}
	return utxoCache.db.Del(pubKeyHash)
}
