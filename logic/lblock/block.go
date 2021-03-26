package lblock

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"reflect"

	"github.com/dappley/go-dappley/core/scState"
	"github.com/dappley/go-dappley/core/transaction"
	"github.com/dappley/go-dappley/logic/ltransaction"
	"github.com/dappley/go-dappley/logic/lutxo"

	"github.com/dappley/go-dappley/vm"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/common/hash"
	"github.com/dappley/go-dappley/core/block"
	"github.com/dappley/go-dappley/crypto/keystore/secp256k1"
	"github.com/dappley/go-dappley/crypto/sha3"
	"github.com/dappley/go-dappley/util"
	logger "github.com/sirupsen/logrus"
)

func HashTransactions(b *block.Block) []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.GetTransactions() {
		txHashes = append(txHashes, tx.Hash())
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func CalculateHash(b *block.Block) hash.Hash {
	return CalculateHashWithNonce(b)
}

func CalculateHashWithoutNonce(b *block.Block) hash.Hash {
	data := bytes.Join(
		[][]byte{
			b.GetPrevHash(),
			HashTransactions(b),
			util.IntToHex(b.GetTimestamp()),
			[]byte(b.GetProducer()),
		},
		[]byte{},
	)

	hasher := sha3.New256()
	hasher.Write(data)
	return hasher.Sum(nil)
}

func CalculateHashWithNonce(b *block.Block) hash.Hash {
	data := bytes.Join(
		[][]byte{
			b.GetPrevHash(),
			HashTransactions(b),
			util.IntToHex(b.GetTimestamp()),
			//util.IntToHex(targetBits),
			util.IntToHex(b.GetNonce()),
			[]byte(b.GetProducer()),
		},
		[]byte{},
	)
	h := sha256.Sum256(data)
	return h[:]
}

func SignBlock(b *block.Block, key string) bool {
	if len(key) <= 0 {
		logger.Warn("Block: the key is too short for signature!")
		return false
	}

	signature, err := generateSignature(key, b.GetHash())
	if err != nil {
		return false
	}
	b.SetSignature(signature)
	return true
}

func generateSignature(key string, data hash.Hash) (hash.Hash, error) {
	privData, err := hex.DecodeString(key)

	if err != nil {
		logger.Warn("Block: cannot decode private key for signature!")
		return []byte{}, err
	}
	signature, err := secp256k1.Sign(data, privData)
	if err != nil {
		logger.WithError(err).Warn("Block: failed to calculate signature!")
		return []byte{}, err
	}

	return signature, nil
}

func VerifyHash(b *block.Block) bool {
	return bytes.Compare(b.GetHash(), CalculateHash(b)) == 0
}

func VerifyTransactions(b *block.Block, utxoIndex *lutxo.UTXOIndex, scState *scState.ScState, parentBlk *block.Block) bool {
	if len(b.GetTransactions()) == 0 {
		logger.WithFields(logger.Fields{
			"hash":   b.GetHash(),
			"height": b.GetHeight(),
		}).Warn("Block: there is no transaction to verify in this block.")
		return false
	}

	var coinbaseTx *transaction.Transaction
	totalTip := common.NewAmount(0)
	totalGasFee := common.NewAmount(0)
	var actualGasList []uint64
	var rewardTX *transaction.Transaction
	// originContractGenTxs: generated by contract in tx list
	var originContractGenTxs []*transaction.Transaction
	rewards := make(map[string]string)
	// currentContractGenTXs: generated by contract execution in validation process
	var currentContractGenTXs []*transaction.Transaction

	scEngine := vm.NewV8Engine()
	defer scEngine.DestroyEngine()
L:
	for _, tx := range b.GetTransactions() {
		totalTip = totalTip.Add(tx.Tip)
		// Collect the contract-incurred transactions in this block
		adaptedTx := transaction.NewTxAdapter(tx)
		if adaptedTx.IsRewardTx() {
			if rewardTX != nil {
				logger.WithFields(logger.Fields{
					"hash":   b.GetHash(),
					"height": b.GetHeight(),
				}).Warn("Block: contains more than 1 reward transaction.")
				return false
			}
			rewardTX = tx
			if !utxoIndex.UpdateUtxo(tx) {
				logger.Warn("VerifyTransactions warn")
			}
			continue L
		}
		if adaptedTx.IsContractSend() {
			originContractGenTxs = append(originContractGenTxs, tx)
		}

		if err := ltransaction.VerifyTransaction(utxoIndex, tx, b.GetHeight()); err != nil {
			logger.WithFields(logger.Fields{
				"hash":   b.GetHash(),
				"height": b.GetHeight(),
			}).Warn(err.Error())
			return false
		}

		ctx := ltransaction.NewTxContract(tx)
		if ctx != nil {
			// Run the contract and collect generated transactions
			gasCount, generatedTxs, err := ltransaction.VerifyAndCollectContractOutput(utxoIndex, ctx, scState, scEngine, b.GetHeight(), parentBlk, rewards)
			if err != nil {
				logger.WithFields(logger.Fields{
					"hash":   b.GetHash(),
					"height": b.GetHeight(),
				}).Warn(err.Error())
				return false
			}
			if generatedTxs != nil {
				currentContractGenTXs = append(currentContractGenTXs, generatedTxs...)
			}
			totalGasFee = totalGasFee.Add(tx.GasLimit.Mul(tx.GasPrice))
			actualGasList = append(actualGasList, gasCount*tx.GasPrice.Uint64())
		} else {
			// tx is a normal transactions
			if !utxoIndex.UpdateUtxo(tx) {
				logger.Warn("VerifyTransactions warn.")
			}
		}
		if adaptedTx.IsCoinbase() {
			if coinbaseTx != nil {
				logger.WithFields(logger.Fields{
					"hash":   b.GetHash(),
					"height": b.GetHeight(),
				}).Warn("Block: contains more than 1 coinbase transaction.")
				return false
			}
			coinbaseTx = tx
		}
	}
	if coinbaseTx == nil {
		logger.WithFields(logger.Fields{
			"hash":   b.GetHash(),
			"height": b.GetHeight(),
		}).Warn("Block: missing coinbase tx.")
		return false
	} else {
		coinbaseAmount := coinbaseTx.Vout[0].Value
		if coinbaseAmount == nil || coinbaseAmount.Cmp(transaction.Subsidy.Add(totalTip)) != 0 {
			logger.WithFields(logger.Fields{
				"hash":   b.GetHash(),
				"height": b.GetHeight(),
			}).Warn("Block: coinbase reward is not right.")
			return false
		}
	}

	if !verifyGasTxs(b.GetTransactions(), totalGasFee, actualGasList) {
		logger.WithFields(logger.Fields{
			"hash":   b.GetHash(),
			"height": b.GetHeight(),
		}).Warn("Block: txs with gas cannot be verified.")
		return false
	}

	// Assert that any contract-incurred transactions matches the ones generated from contract execution
	if rewardTX != nil && !rewardTX.MatchRewards(rewards) {
		logger.WithFields(logger.Fields{
			"hash":   b.GetHash(),
			"height": b.GetHeight(),
		}).Warn("Block: reward tx cannot be verified.")
		return false
	}
	if len(originContractGenTxs) > 0 && !verifyGeneratedTXs(utxoIndex, originContractGenTxs, currentContractGenTXs) {
		logger.WithFields(logger.Fields{
			"hash":   b.GetHash(),
			"height": b.GetHeight(),
		}).Warn("Block: generated tx cannot be verified.")
		return false
	}
	return true
}

// verifyGeneratedTXs verify that transactions generated by gas reward or change is same with its inputs
func verifyGasTxs(blockTxs []*transaction.Transaction, totalGasFee *common.Amount, actualGasList []uint64) bool {
	if totalGasFee.IsZero() && len(actualGasList) == 0 {
		return true
	}
	var err error
	for _, tx := range blockTxs {
		adaptedTx := transaction.NewTxAdapter(tx)
		if adaptedTx.IsGasRewardTx() {
			txGasReward := ltransaction.TxGasReward{tx}
			rewardValue := txGasReward.GetRewardValue()
			if rewardValue == nil {
				return false
			}
			isFound := false
			for i, gasCount := range actualGasList {
				if gasCount == rewardValue.Uint64() {
					actualGasList = append(actualGasList[:i], actualGasList[i+1:]...)
					totalGasFee, err = totalGasFee.Sub(rewardValue)
					if err != nil {
						return false
					}
					isFound = true
					break
				}
			}
			if !isFound {
				return false
			}
		} else if adaptedTx.IsGasChangeTx() {
			txGasChange := ltransaction.TxGasChange{tx}
			changedValue := txGasChange.GetChangeValue()
			if changedValue == nil {
				return false
			}
			totalGasFee, err = totalGasFee.Sub(changedValue)
			if err != nil {
				return false
			}
		}
	}
	if !totalGasFee.IsZero() {
		return false
	}
	return true
}

// verifyGeneratedTXs verify that all transactions in candidates can be found in generatedTXs
func verifyGeneratedTXs(utxoIndex *lutxo.UTXOIndex, candidates []*transaction.Transaction, generatedTXs []*transaction.Transaction) bool {
	// genTXBuckets stores description of txs grouped by concatenation of sender's and recipient's public key hashes
	genTXBuckets := make(map[string][][]*common.Amount)
	for _, genTX := range generatedTXs {
		sender, recipient, amount, tip, err := ltransaction.DescribeTransaction(utxoIndex, genTX)
		if err != nil {
			continue
		}
		hashKey := sender.String() + recipient.String()
		genTXBuckets[hashKey] = append(genTXBuckets[hashKey], []*common.Amount{amount, tip})
	}
L:
	for _, tx := range candidates {
		sender, recipient, amount, tip, err := ltransaction.DescribeTransaction(utxoIndex, tx)
		if err != nil {
			return false
		}
		hashKey := sender.String() + recipient.String()
		if genTXBuckets[hashKey] == nil {
			return false
		}
		for i, t := range genTXBuckets[hashKey] {
			// tx is verified if amount and tip matches
			if amount.Cmp(t[0]) == 0 && tip.Cmp(t[1]) == 0 {
				genTXBuckets[hashKey] = append(genTXBuckets[hashKey][:i], genTXBuckets[hashKey][i+1:]...)
				continue L
			}
		}
		return false
	}
	return true
}

func IsHashEqual(h1 hash.Hash, h2 hash.Hash) bool {

	return reflect.DeepEqual(h1, h2)
}
