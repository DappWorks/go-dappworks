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
	"bytes"
	"errors"
	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core/account"
	utxopb "github.com/dappley/go-dappley/core/utxo/pb"
	"github.com/dappley/go-dappley/storage"
	"github.com/dappley/go-dappley/util"
	"github.com/golang/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	logger "github.com/sirupsen/logrus"
)

const UtxoCacheLRUCacheLimit = 1024

// UTXOCache holds temporary data
type UTXOCache struct {
	contractCreateCache *lru.Cache
	cache               *lru.Cache
	utxo                *lru.Cache
	utxoInfo            *lru.Cache
	db                  storage.Storage
}

func NewUTXOCache(db storage.Storage) *UTXOCache {
	utxoCache := &UTXOCache{
		contractCreateCache: nil,
		cache:               nil,
		utxo:                nil,
		utxoInfo:            nil,
		db:                  db,
	}
	utxoCache.cache, _ = lru.New(UtxoCacheLRUCacheLimit)
	utxoCache.utxo, _ = lru.New(UtxoCacheLRUCacheLimit)
	utxoCache.utxoInfo, _ = lru.New(UtxoCacheLRUCacheLimit)
	utxoCache.contractCreateCache, _ = lru.New(UtxoCacheLRUCacheLimit)
	return utxoCache
}

func (utxoCache *UTXOCache) AddUtxos(utxoTx *UTXOTx, pubkey string) error {
	lastestUtxoKey := utxoCache.getLastUTXOKey(pubkey)
	for key, utxo := range utxoTx.Indices {
		if bytes.Equal(util.Str2bytes(key), lastestUtxoKey) {
			return errors.New("add utxo failed: the utxo is same as the last utxo")
		}
		utxo.NextUtxoKey = lastestUtxoKey
		err := utxoCache.putUTXOToDB(utxo)
		if err != nil {
			return err
		}
		lastestUtxoKey = util.Str2bytes(key)

		if utxo.UtxoType == UtxoCreateContract {
			err := utxoCache.putCreateContractUTXOKey(pubkey, util.Str2bytes(key))
			if err != nil {
				return err
			}
		}
	}
	err := utxoCache.putLastUTXOKey(pubkey, lastestUtxoKey)
	if err != nil {
		return err
	}
	return nil
}

func (utxoCache *UTXOCache) RemoveUtxos(utxoTx *UTXOTx, pubkey string) error {
	for key, utxo := range utxoTx.Indices {
		preUTXO, err := utxoCache.GetPreUtxo(pubkey, key)
		if err != nil {
			return err
		}
		if preUTXO == nil { //this utxo is the head utxo
			if bytes.Equal(utxo.NextUtxoKey, []byte{}) {
				err = utxoCache.deleteUTXOInfo(pubkey)
				if err != nil {
					return err
				}
			} else {
				err := utxoCache.putLastUTXOKey(pubkey, utxo.NextUtxoKey)
				if err != nil {
					return err
				}
			}
		} else {
			if bytes.Equal(utxo.NextUtxoKey, util.Str2bytes(preUTXO.GetUTXOKey())) {
				return errors.New("remove utxo error: find duplicate utxo in db")
			}
			preUTXO.NextUtxoKey = utxo.NextUtxoKey
			err = utxoCache.putUTXOToDB(preUTXO)
			if err != nil {
				return err
			}
			//*if this utxo in utxoTx,then need to update
			if _, ok := utxoTx.Indices[preUTXO.GetUTXOKey()]; ok {
				utxoTx.Indices[preUTXO.GetUTXOKey()] = preUTXO
			}
		}
		err = utxoCache.deleteUTXOFromDB(key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (utxoCache *UTXOCache) GetUtxo(utxoKey string) (*UTXO, error) {
	utxoData, ok := utxoCache.utxo.Get(utxoKey)
	if ok {
		utxo := utxoData.(*UTXO)
		return utxo, nil
	}
	return utxoCache.getUTXOFromDB(utxoKey)
}

func (utxoCache *UTXOCache) GetPreUtxo(pubKey, targetUtxokey string) (*UTXO, error) {
	utxoKey := utxoCache.getLastUTXOKey(pubKey)
	if bytes.Equal(utxoKey, []byte{}) {
		return nil, errors.New("this pubkey's utxo is not exist")
	}
	if bytes.Equal(utxoKey, util.Str2bytes(targetUtxokey)){
		return nil, nil
	}
	for !bytes.Equal(utxoKey, []byte{}) {
		utxo, err := utxoCache.GetUtxo(util.Bytes2str(utxoKey))
		if err != nil {
			return nil, err
		}
		if bytes.Equal(utxo.NextUtxoKey, util.Str2bytes(targetUtxokey)) {
			return utxo, nil
		}
		utxoKey = utxo.NextUtxoKey
	}
	return nil, nil
}

func (utxoCache *UTXOCache) getLastUTXOKey(pubKeyHash string) []byte {
	utxoInfo, err := utxoCache.getUTXOInfo(pubKeyHash)
	if err != nil {
		logger.Warn("getLastUTXOKey error:", err)
		return []byte{}
	}
	return utxoInfo.GetLastUtxoKey()
}

func (utxoCache *UTXOCache) IsLastUtxoKeyExist(pubKeyHash string) bool {
	if bytes.Equal(utxoCache.getLastUTXOKey(pubKeyHash), []byte{}) {
		return false
	}
	return true
}

func (utxoCache *UTXOCache) GetUTXOTx(pubKeyHash account.PubKeyHash) *UTXOTx {
	lastUtxokey := utxoCache.getLastUTXOKey(pubKeyHash.String())
	utxoTx := NewUTXOTx()
	utxoKey := util.Bytes2str(lastUtxokey)
	//count:=0
	for utxoKey != "" {
		utxo, err := utxoCache.GetUtxo(utxoKey)
		if err != nil {
			//todo: return err
			logger.Error("GetUTXOTx:", err)
		}
		utxoTx.Indices[utxoKey] = utxo
		utxoKey = util.Bytes2str(utxo.NextUtxoKey) //get previous utxo key
		//count++
		//if count==1000{
		//	break
		//}
	}
	return &utxoTx
}

func (utxoCache *UTXOCache) GetUTXOsByAmountWithOutRemovedUTXOs(pubKeyHash account.PubKeyHash,amount *common.Amount, utxoTxRemove *UTXOTx) ([]*UTXO,error) {
	lastUtxokey := utxoCache.getLastUTXOKey(pubKeyHash.String())
	var utxoSlice []*UTXO
	targetAmount := common.NewAmount(0)
	utxoKey := util.Bytes2str(lastUtxokey)

	for utxoKey != "" {
		utxo, err := utxoCache.GetUtxo(utxoKey)
		if err != nil {
			logger.Warn( err)
		}
		if utxoTxRemove != nil {
			if _, ok := utxoTxRemove.Indices[utxo.GetUTXOKey()]; ok {
				delete(utxoTxRemove.Indices, utxo.GetUTXOKey())
				continue
			}
		}
		targetAmount = targetAmount.Add(utxo.Value)
		utxoSlice = append(utxoSlice, utxo)
		if targetAmount.Cmp(amount) >= 0 {
			return utxoSlice, nil
		}

		utxoKey = util.Bytes2str(utxo.NextUtxoKey) //get previous utxo key
	}
	return nil, errors.New("transaction: insufficient balance")
}

func (utxoCache *UTXOCache) putUTXOToDB(utxo *UTXO) error {
	utxoBytes, err := proto.Marshal(utxo.ToProto().(*utxopb.Utxo))
	if err != nil {
		return err
	}
	err = utxoCache.db.Put(util.Str2bytes(utxo.GetUTXOKey()), utxoBytes)
	if err != nil {
		logger.Error("put utxo to db failed！")
		return err
	}
	utxoCache.utxo.Add(utxo.GetUTXOKey(), utxo)
	return nil
}

func (utxoCache *UTXOCache) getUTXOFromDB(utxoKey string) (*UTXO, error) {
	var utxo = &UTXO{}
	rawBytes, err := utxoCache.db.Get(util.Str2bytes(utxoKey))
	if err == nil {
		utxoPb := &utxopb.Utxo{}
		err := proto.Unmarshal(rawBytes, utxoPb)
		if err != nil {
			logger.Error("Unmarshal utxo failed.")
			return nil, err
		}
		utxo.FromProto(utxoPb)
		utxoCache.utxo.Add(utxoKey, utxo)
		return utxo, nil
	}
	logger.Error("get utxo from db failed！")
	return nil, err
}

func (utxoCache *UTXOCache) deleteUTXOFromDB(utxoKey string) error {
	err := utxoCache.db.Del(util.Str2bytes(utxoKey))
	if err != nil {
		logger.Error("delete utxo from db failed.")
		return err
	}
	utxoCache.utxo.Remove(utxoKey)
	return nil
}

func (utxoCache *UTXOCache) putLastUTXOKey(pubkey string, lastUTXOKey []byte) error {
	utxoInfo, err := utxoCache.getUTXOInfo(pubkey)
	if err != nil {
		logger.Warn("putLastUTXOKey:", err)
	}
	utxoInfo.SetLastUtxoKey(lastUTXOKey)
	err = utxoCache.putUTXOInfo(pubkey, utxoInfo)
	if err != nil {
		logger.Error("put last utxo key to db failed.")
		return err
	}
	return nil
}

func (utxoCache *UTXOCache) getUTXOInfo(pubkey string) (*UTXOInfo, error) {
	utxoInfoData, ok := utxoCache.utxoInfo.Get(pubkey)
	if ok {
		return utxoInfoData.(*UTXOInfo), nil
	}

	utxoInfo := NewUTXOInfo()
	rawBytes, err := utxoCache.db.Get(util.Str2bytes(pubkey))
	if err != nil {
		logger.Warn("utxoInfo not found in db")
		return utxoInfo, err
	}

	utxoInfoPb := &utxopb.UtxoInfo{}
	err = proto.Unmarshal(rawBytes, utxoInfoPb)
	if err != nil {
		logger.Error("Unmarshal utxo info failed.")
		return utxoInfo, err
	}
	utxoInfo.FromProto(utxoInfoPb)

	utxoCache.utxoInfo.Add(pubkey, utxoInfo)
	return utxoInfo, nil
}

func (utxoCache *UTXOCache) putUTXOInfo(pubkey string, utxoInfo *UTXOInfo) error {
	utxoBytes, err := proto.Marshal(utxoInfo.ToProto().(*utxopb.UtxoInfo))
	if err != nil {
		return err
	}
	err = utxoCache.db.Put(util.Str2bytes(pubkey), utxoBytes)
	if err != nil {
		logger.Error("put utxoInfo to db failed.")
		return err
	}
	utxoCache.utxoInfo.Add(pubkey, utxoInfo)
	return nil
}

func (utxoCache *UTXOCache) deleteUTXOInfo(pubkey string) error {
	err := utxoCache.db.Del(util.Str2bytes(pubkey))
	if err != nil {
		logger.Error("deleteUTXOInfo: delete utxoInfo from db failed.")
		return err
	}
	utxoCache.utxoInfo.Remove(pubkey)
	return nil
}

func (utxoCache *UTXOCache) putCreateContractUTXOKey(pubkey string, createContractUTXOKey []byte) error {
	utxoInfo, err := utxoCache.getUTXOInfo(pubkey)
	if err != nil {
		logger.Warn(err)
	}
	utxoInfo.SetCreateContractUTXOKey(createContractUTXOKey)
	err = utxoCache.putUTXOInfo(pubkey, utxoInfo)
	if err != nil {
		logger.Error("put utxoCreateContractKey to db failed.")
		return err
	}
	return nil
}

func (utxoCache *UTXOCache) GetUtxoCreateContract(pubKeyHash string) *UTXO {
	utxoInfo, err := utxoCache.getUTXOInfo(pubKeyHash)
	if err != nil || utxoInfo.GetCreateContractUTXOKey() == nil {
		return nil
	}
	utxo, err := utxoCache.GetUtxo(util.Bytes2str(utxoInfo.GetCreateContractUTXOKey()))
	if err != nil {
		logger.Error("Get UtxoCreateContract failed.")
		return nil
	}
	return utxo
}
