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
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core/pb"
	"github.com/dappley/go-dappley/crypto/byteutils"
	"github.com/dappley/go-dappley/crypto/keystore/secp256k1"
	"github.com/dappley/go-dappley/util"
	"github.com/golang/protobuf/proto"
	logger "github.com/sirupsen/logrus"
)

var subsidy = common.NewAmount(10000000)

const ContractTxouputIndex = 0

var rewardTxData = []byte("Distribute X Rewards")

var (
	// MinGasCountPerTransaction default gas for normal transaction
	MinGasCountPerTransaction = common.NewAmount(20000)
	// GasCountPerByte per byte of data attached to a transaction gas cost
	GasCountPerByte = common.NewAmount(1)

	ErrInsufficientFund  = errors.New("transaction: insufficient balance")
	ErrInvalidAmount     = errors.New("transaction: invalid amount (must be > 0)")
	ErrTXInputNotFound   = errors.New("transaction: transaction input not found")
	ErrNewUserPubKeyHash = errors.New("transaction: create pubkeyhash error")
	ErrNoGasChange       = errors.New("transaction: all of Gas have been consumed")
)

type Transaction struct {
	ID       []byte
	Vin      []TXInput
	Vout     []TXOutput
	Tip      *common.Amount
	GasPrice *common.Amount
	GasLimit *common.Amount
}

// ContractTx contains contract value
type ContractTx struct {
	Transaction
}

type TxIndex struct {
	BlockId    []byte
	BlockIndex int
}

// Transaction parameters
type SendTxParam struct {
	From          Address
	SenderKeyPair *KeyPair
	To            Address
	Amount        *common.Amount
	Tip           *common.Amount
	GasLimit      *common.Amount
	GasPrice      *common.Amount
	Contract      string
}

// Returns SendTxParam object
func NewSendTxParam(from Address, senderKeyPair *KeyPair, to Address, amount *common.Amount, tip *common.Amount, gasLimit *common.Amount, gasPrice *common.Amount, contract string) SendTxParam {
	return SendTxParam{from, senderKeyPair, to, amount, tip, gasLimit, gasPrice, contract}
}

// TotalCost returns total cost of utxo value in this transaction
func (st SendTxParam) TotalCost() *common.Amount {
	var totalAmount = st.Amount
	if st.Tip != nil {
		totalAmount = totalAmount.Add(st.Tip)
	}
	if st.GasLimit != nil {
		limitedFee := st.GasLimit.Mul(st.GasPrice)
		totalAmount = totalAmount.Add(limitedFee)
	}
	return totalAmount
}

// Returns structure of ContractTx
func (tx *Transaction) ToContractTx() *ContractTx {
	if !tx.IsContract() {
		return nil
	}
	return &ContractTx{*tx}
}

func (tx *Transaction) IsCoinbase() bool {

	if !tx.isVinCoinbase() {
		return false
	}

	if len(tx.Vout) != 1 {
		return false
	}

	if bytes.Equal(tx.Vin[0].PubKey, rewardTxData) {
		return false
	}

	return true
}

func (tx *Transaction) IsRewardTx() bool {

	if !tx.isVinCoinbase() {
		return false
	}

	if !bytes.Equal(tx.Vin[0].PubKey, rewardTxData) {
		return false
	}

	return true
}

// IsContract returns true if tx deploys/executes a smart contract; false otherwise
func (tx *Transaction) IsContract() bool {
	if len(tx.Vout) == 0 {
		return false
	}
	isContract, _ := tx.Vout[ContractTxouputIndex].PubKeyHash.IsContract()
	return isContract
}

func (ctx *ContractTx) IsContract() bool {
	return true
}

func (ctx *ContractTx) IsExecutionContract() bool {
	if !strings.Contains(ctx.GetContract(), scheduleFuncName) {
		return true
	}
	return false
}

// IsFromContract returns true if tx is generated from a contract execution; false otherwise
func (tx *Transaction) IsFromContract(utxoIndex *UTXOIndex) bool {
	if len(tx.Vin) == 0 {
		return false
	}

	contractUtxos := utxoIndex.GetContractUtxos()

	for _, vin := range tx.Vin {
		pubKey := PubKeyHash(vin.PubKey)
		if isContract, _ := pubKey.IsContract(); !isContract {
			return false
		}

		if !tx.isPubkeyInUtxos(contractUtxos, pubKey) {
			return false
		}
	}
	return true
}

func (tx *Transaction) isPubkeyInUtxos(contractUtxos []*UTXO, pubKey PubKeyHash) bool {
	for _, contractUtxo := range contractUtxos {
		if bytes.Compare(contractUtxo.PubKeyHash, pubKey) == 0 {
			return true
		}
	}
	return false
}

func (tx *Transaction) isVinCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// Describe reverse-engineers the high-level description of a transaction
func (tx *Transaction) Describe(utxoIndex *UTXOIndex) (sender, recipient *Address, amount, tip *common.Amount, error error) {
	var receiverAddress Address
	vinPubKey := tx.Vin[0].PubKey
	pubKeyHash := PubKeyHash([]byte(""))
	inputAmount := common.NewAmount(0)
	outputAmount := common.NewAmount(0)
	payoutAmount := common.NewAmount(0)
	for _, vin := range tx.Vin {
		if bytes.Compare(vin.PubKey, vinPubKey) == 0 {
			switch {
			case tx.IsRewardTx():
				pubKeyHash = PubKeyHash(rewardTxData)
				continue
			case tx.IsFromContract(utxoIndex):
				// vinPubKey is the pubKeyHash if it is a sc generated tx
				pubKeyHash = PubKeyHash(vinPubKey)
			default:
				pkh, err := NewUserPubKeyHash(vin.PubKey)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				pubKeyHash = pkh
			}
			usedUTXO := utxoIndex.FindUTXOByVin([]byte(pubKeyHash), vin.Txid, vin.Vout)
			inputAmount = inputAmount.Add(usedUTXO.Value)
		} else {
			logger.Debug("Transaction: using UTXO from multiple wallets.")
		}
	}
	for _, vout := range tx.Vout {
		if bytes.Compare([]byte(vout.PubKeyHash), vinPubKey) == 0 {
			outputAmount = outputAmount.Add(vout.Value)
		} else {
			receiverAddress = vout.PubKeyHash.GenerateAddress()
			payoutAmount = payoutAmount.Add(vout.Value)
		}
	}
	tip, err := inputAmount.Sub(outputAmount)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	senderAddress := pubKeyHash.GenerateAddress()

	return &senderAddress, &receiverAddress, payoutAmount, tip, nil
}

//GetToHashBytes Get bytes for hash
func (tx *Transaction) GetToHashBytes() []byte {
	var tempBytes []byte

	for _, vin := range tx.Vin {
		tempBytes = bytes.Join([][]byte{
			tempBytes,
			vin.Txid,
			byteutils.FromInt32(int32(vin.Vout)),
			vin.PubKey,
			vin.Signature,
		}, []byte{})
	}

	for _, vout := range tx.Vout {
		tempBytes = bytes.Join([][]byte{
			tempBytes,
			vout.Value.Bytes(),
			[]byte(vout.PubKeyHash),
			[]byte(vout.Contract),
		}, []byte{})
	}
	if tx.Tip != nil {
		tempBytes = append(tempBytes, tx.Tip.Bytes()...)
	}
	return tempBytes
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	hash = sha256.Sum256(tx.GetToHashBytes())

	return hash[:]
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevUtxos []*UTXO) error {
	if tx.IsCoinbase() {
		logger.Warn("Transaction: will not sign a coinbase transaction.")
		return nil
	}

	if tx.IsRewardTx() {
		logger.Warn("Transaction: will not sign a reward transaction.")
		return nil
	}

	txCopy := tx.TrimmedCopy(false)
	privData, err := secp256k1.FromECDSAPrivateKey(&privKey)
	if err != nil {
		logger.WithError(err).Error("Transaction: failed to get private key.")
		return err
	}

	for i, vin := range txCopy.Vin {
		txCopy.Vin[i].Signature = nil
		oldPubKey := vin.PubKey
		txCopy.Vin[i].PubKey = []byte(prevUtxos[i].PubKeyHash)
		txCopy.ID = txCopy.Hash()

		txCopy.Vin[i].PubKey = oldPubKey

		signature, err := secp256k1.Sign(txCopy.ID, privData)
		if err != nil {
			logger.WithError(err).Error("Transaction: failed to create a signature.")
			return err
		}

		tx.Vin[i].Signature = signature
	}
	return nil
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy(withSignature bool) Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	var pubkey []byte

	for _, vin := range tx.Vin {
		if withSignature {
			pubkey = vin.PubKey
		} else {
			pubkey = nil
		}
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, pubkey})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash, vout.Contract})
	}

	txCopy := Transaction{tx.ID, inputs, outputs, tx.Tip, tx.GasLimit, tx.GasPrice}

	return txCopy
}

func (tx *Transaction) DeepCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, vin.Signature, vin.PubKey})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash, vout.Contract})
	}

	txCopy := Transaction{tx.ID, inputs, outputs, tx.Tip, tx.GasLimit, tx.GasPrice}

	return txCopy
}

func (ctx *ContractTx) IsContractDeployed(utxoIndex *UTXOIndex) bool {
	pubkeyhash := ctx.GetContractPubKeyHash()
	if pubkeyhash == nil {
		return false
	}

	contractUtxoTx := utxoIndex.GetAllUTXOsByPubKeyHash(pubkeyhash)
	return contractUtxoTx.Size() > 0
}

// Verify ensures signature of transactions is correct or verifies against blockHeight if it's a coinbase transactions
func (tx *Transaction) Verify(utxoIndex *UTXOIndex, blockHeight uint64) error {
	ctx := tx.ToContractTx()
	if ctx != nil {
		logger.Warn("Verify ctx.")
		return ctx.Verify(utxoIndex)
	}
	logger.Warn("Verify not ctx.")
	if tx.IsCoinbase() {
		//TODO coinbase vout check need add tip
		if tx.Vout[0].Value.Cmp(subsidy) < 0 {
			logger.Warn("Transaction: subsidy check failed.")
			return errors.New("Transaction: subsidy check failed")
		}
		bh := binary.BigEndian.Uint64(tx.Vin[0].Signature)
		if blockHeight != bh {
			logger.Warnf("Transaction: block height check failed expected=%v actual=%v.", blockHeight, bh)
			return fmt.Errorf("Transaction: block height check failed expected=%v actual=%v.", blockHeight, bh)
		}
		return nil
	}
	if tx.IsRewardTx() {
		//TODO: verify reward tx here
		return nil
	}

	_, err := verify(tx, utxoIndex)
	return err
}

// Verify ensures signature of transactions is correct or verifies against blockHeight if it's a coinbase transactions
func (ctx *ContractTx) Verify(utxoIndex *UTXOIndex) error {
	logger.Warn("IsExecutionContract.", ctx.IsExecutionContract())
	logger.Warn("IsContractDeployed.", ctx.IsContractDeployed(utxoIndex))
	if ctx.IsExecutionContract() && !ctx.IsContractDeployed(utxoIndex) {
		logger.Warn("Transaction: contract state check failed.")
		return errors.New("Transaction: contract state check failed")
	}

	totalBalance, err := verify(&ctx.Transaction, utxoIndex)
	if err != nil {
		return err
	}
	return ctx.verifyGas(totalBalance)
}

// VerifyInEstimate returns whether the current tx in estimate mode is valid.
func (ctx *ContractTx) VerifyInEstimate(utxoIndex *UTXOIndex) error {
	if ctx.IsExecutionContract() && !ctx.IsContractDeployed(utxoIndex) {
		return errors.New("Transaction: contract state check failed")
	}

	_, err := verify(&ctx.Transaction, utxoIndex)
	if err != nil {
		return err
	}
	return nil
}

func verify(tx *Transaction, utxoIndex *UTXOIndex) (*common.Amount, error) {
	prevUtxos := getPrevUTXOs(tx, utxoIndex)
	if prevUtxos == nil {
		return nil, errors.New("Transaction: prevUtxos not found")
	}

	if !tx.verifyID() {
		logger.WithFields(logger.Fields{
			"tx_id": hex.EncodeToString(tx.ID),
		}).Warn("Transaction: ID is invalid.")
		return nil, errors.New("Transaction: ID is invalid")
	}

	if !tx.verifyPublicKeyHash(prevUtxos) {
		logger.WithFields(logger.Fields{
			"tx_id": hex.EncodeToString(tx.ID),
		}).Warn("Transaction: pubkey is invalid.")
		return nil, errors.New("Transaction: pubkey is invalid")
	}

	totalPrev := calculateUtxoSum(prevUtxos)
	totalVoutValue, ok := tx.calculateTotalVoutValue()
	if !ok {
		return nil, errors.New("Transaction: vout is invalid")
	}

	if !tx.verifyAmount(totalPrev, totalVoutValue) {
		logger.WithFields(logger.Fields{
			"tx_id": hex.EncodeToString(tx.ID),
		}).Warn("Transaction: amount is invalid.")
		return nil, errors.New("Transaction: amount is invalid")
	}

	if !tx.verifyTip(totalPrev, totalVoutValue) {
		logger.WithFields(logger.Fields{
			"tx_id": hex.EncodeToString(tx.ID),
		}).Warn("Transaction: tip is invalid.")
		return nil, errors.New("Transaction: tip is invalid")
	}

	if !tx.verifySignatures(prevUtxos) {
		logger.WithFields(logger.Fields{
			"tx_id": hex.EncodeToString(tx.ID),
		}).Warn("Transaction: signature is invalid.")
		return nil, errors.New("Transaction: signature is invalid")
	}
	totalBalance, _ := totalPrev.Sub(totalVoutValue)
	totalBalance, _ = totalBalance.Sub(tx.Tip)
	logger.WithFields(logger.Fields{
		"totalBalance": totalBalance,
	}).Warn("Transaction: totalBalance.")
	return totalBalance, nil
}

// Returns related previous UTXO for current transaction
func getPrevUTXOs(tx *Transaction, utxoIndex *UTXOIndex) []*UTXO {
	var prevUtxos []*UTXO
	tempUtxoTxMap := make(map[string]*UTXOTx)
	for _, vin := range tx.Vin {
		pubKeyHash, err := NewUserPubKeyHash(vin.PubKey)
		if err != nil {
			logger.WithFields(logger.Fields{
				"tx_id":          hex.EncodeToString(tx.ID),
				"vin_tx_id":      hex.EncodeToString(vin.Txid),
				"vin_public_key": hex.EncodeToString(vin.PubKey),
			}).Warn("Transaction: failed to get PubKeyHash of vin.")
			return nil
		}
		tempUtxoTx, ok := tempUtxoTxMap[string(pubKeyHash)]
		if !ok {
			tempUtxoTx = utxoIndex.GetAllUTXOsByPubKeyHash(pubKeyHash)
			tempUtxoTxMap[string(pubKeyHash)] = tempUtxoTx
		}
		utxo := tempUtxoTx.GetUtxo(vin.Txid, vin.Vout)
		if utxo == nil {
			logger.WithFields(logger.Fields{
				"tx_id":      hex.EncodeToString(tx.ID),
				"vin_tx_id":  hex.EncodeToString(vin.Txid),
				"vin_index":  vin.Vout,
				"pubKeyHash": hex.EncodeToString(pubKeyHash),
			}).Warn("Transaction: cannot find vin.")
			return nil
		}
		prevUtxos = append(prevUtxos, utxo)
	}
	return prevUtxos
}

// verifyID verifies if the transaction ID is the hash of the transaction
func (tx *Transaction) verifyID() bool {
	txCopy := tx.TrimmedCopy(true)
	if bytes.Equal(tx.ID, (&txCopy).Hash()) {
		return true
	} else {
		return false
	}
}

// verifyAmount verifies if the transaction has the correct vout value
func (tx *Transaction) verifyAmount(totalPrev *common.Amount, totalVoutValue *common.Amount) bool {
	//TotalVin amount must equal or greater than total vout
	return totalPrev.Cmp(totalVoutValue) >= 0
}

//verifyTip verifies if the transaction has the correct tip
func (tx *Transaction) verifyTip(totalPrev *common.Amount, totalVoutValue *common.Amount) bool {
	logger.WithFields(logger.Fields{
		"totalPrev":      totalPrev,
		"totalVoutValue": totalVoutValue,
	}).Error("verifyTip.")
	sub, err := totalPrev.Sub(totalVoutValue)
	if err != nil {
		return false
	}
	return tx.Tip.Cmp(sub) <= 0
}

// verifyGas verifies if the transaction has the correct GasLimit and GasPrice
func (ctx *ContractTx) verifyGas(totalBalance *common.Amount) error {
	baseGas, err := ctx.GasCountOfTxBase()
	if err == nil {
		logger.WithFields(logger.Fields{
			"baseGas":      baseGas,
			"ctx.GasLimit": ctx.GasLimit,
		}).Warn("verifyGas: baseGas.")
		if ctx.GasLimit.Cmp(baseGas) < 0 {
			logger.WithFields(logger.Fields{
				"error":       ErrOutOfGasLimit,
				"transaction": ctx,
				"limit":       ctx.GasLimit,
				"acceptedGas": baseGas,
			}).Warn("Failed to check GasLimit >= txBaseGas.")
			// GasLimit is smaller than based tx gas, won't giveback the tx
			return ErrOutOfGasLimit
		}
	}

	limitedFee := ctx.GasLimit.Mul(ctx.GasPrice)
	logger.WithFields(logger.Fields{
		"limitedFee": limitedFee,
	}).Warn("verifyGas: limitedFee.")
	if totalBalance.Cmp(limitedFee) < 0 {
		return ErrInsufficientBalance
	}
	return nil
}

//calculateTotalVoutValue returns total amout of transaction's vout
func (tx *Transaction) calculateTotalVoutValue() (*common.Amount, bool) {
	totalVout := &common.Amount{}
	for _, vout := range tx.Vout {
		if vout.Value == nil || vout.Value.Validate() != nil {
			return nil, false
		}
		totalVout = totalVout.Add(vout.Value)
	}
	return totalVout, true
}

//verifyPublicKeyHash verifies if the public key in Vin is the original key for the public
//key hash in utxo
func (tx *Transaction) verifyPublicKeyHash(prevUtxos []*UTXO) bool {

	for i, vin := range tx.Vin {
		if prevUtxos[i].PubKeyHash == nil {
			logger.Error("Transaction: previous transaction is not correct.")
			return false
		}

		isContract, err := prevUtxos[i].PubKeyHash.IsContract()
		if err != nil {
			return false
		}
		//if the utxo belongs to a Contract, the utxo is not verified through
		//public key hash. It will be verified through consensus
		if isContract {
			continue
		}

		pubKeyHash, err := NewUserPubKeyHash(vin.PubKey)
		if err != nil {
			return false
		}
		if !bytes.Equal([]byte(pubKeyHash), []byte(prevUtxos[i].PubKeyHash)) {
			return false
		}
	}
	return true
}

func (tx *Transaction) verifySignatures(prevUtxos []*UTXO) bool {
	txCopy := tx.TrimmedCopy(false)

	for i, vin := range tx.Vin {
		txCopy.Vin[i].Signature = nil
		oldPubKey := txCopy.Vin[i].PubKey
		txCopy.Vin[i].PubKey = []byte(prevUtxos[i].PubKeyHash)
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[i].PubKey = oldPubKey

		originPub := make([]byte, 1+len(vin.PubKey))
		originPub[0] = 4 // uncompressed point
		copy(originPub[1:], vin.PubKey)

		verifyResult, err := secp256k1.Verify(txCopy.ID, vin.Signature, originPub)

		if err != nil || verifyResult == false {
			logger.WithError(err).Error("Transaction: signature cannot be verified.")
			return false
		}
	}

	return true
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(to Address, data string, blockHeight uint64, tip *common.Amount) Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	bh := make([]byte, 8)
	binary.BigEndian.PutUint64(bh, uint64(blockHeight))

	txin := TXInput{nil, -1, bh, []byte(data)}
	txout := NewTXOutput(subsidy.Add(tip), to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0)}
	tx.ID = tx.Hash()

	return tx
}

//NewRewardTx creates a new transaction that gives reward to addresses according to the input rewards
func NewRewardTx(blockHeight uint64, rewards map[string]string) Transaction {

	bh := make([]byte, 8)
	binary.BigEndian.PutUint64(bh, uint64(blockHeight))

	txin := TXInput{nil, -1, bh, rewardTxData}
	txOutputs := []TXOutput{}
	for address, amount := range rewards {
		amt, err := common.NewAmountFromString(amount)
		if err != nil {
			logger.WithError(err).WithFields(logger.Fields{
				"address": address,
				"amount":  amount,
			}).Warn("Transaction: failed to parse reward amount")
		}
		txOutputs = append(txOutputs, *NewTXOutput(amt, NewAddress(address)))
	}
	tx := Transaction{nil, []TXInput{txin}, txOutputs, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0)}
	tx.ID = tx.Hash()

	return tx
}

// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(utxos []*UTXO, sendTxParam SendTxParam) (Transaction, error) {

	sum := calculateUtxoSum(utxos)
	change, err := calculateChange(sum, sendTxParam.Amount, sendTxParam.Tip, sendTxParam.GasLimit, sendTxParam.GasPrice)
	if err != nil {
		return Transaction{}, err
	}
	logger.WithError(err).WithFields(logger.Fields{
		"sum":      sum,
		"Amount":   sendTxParam.Amount,
		"Tip":      sendTxParam.Tip,
		"change":   change,
		"GasLimit": sendTxParam.GasLimit,
		"GasPrice": sendTxParam.GasPrice,
	}).Error("NewUTXOTransaction")
	tx := Transaction{
		nil,
		prepareInputLists(utxos, sendTxParam.SenderKeyPair.PublicKey, nil),
		prepareOutputLists(sendTxParam.From, sendTxParam.To, sendTxParam.Amount, change, sendTxParam.Contract),
		sendTxParam.Tip,
		sendTxParam.GasPrice,
		sendTxParam.GasLimit,
	}
	tx.ID = tx.Hash()

	err = tx.Sign(sendTxParam.SenderKeyPair.PrivateKey, utxos)
	if err != nil {
		return Transaction{}, err
	}

	return tx, nil
}

func NewContractTransferTX(utxos []*UTXO, contractAddr, toAddr Address, amount, tip *common.Amount, gasLimit *common.Amount, gasPrice *common.Amount, sourceTXID []byte) (Transaction, error) {
	contractPubKeyHash, ok := contractAddr.GetPubKeyHash()
	if !ok {
		return Transaction{}, ErrInvalidAddress
	}
	if isContract, err := (PubKeyHash(contractPubKeyHash)).IsContract(); !isContract {
		return Transaction{}, err
	}

	sum := calculateUtxoSum(utxos)
	change, err := calculateChange(sum, amount, tip, gasLimit, gasPrice)
	logger.WithError(err).WithFields(logger.Fields{
		"sum":      sum,
		"Amount":   amount,
		"Tip":      tip,
		"change":   change,
		"gasLimit": gasLimit,
		"gasPrice": gasPrice,
	}).Error("NewContractTransferTX")
	if err != nil {
		return Transaction{}, err
	}

	// Intentionally set PubKeyHash as PubKey (to recognize it is from contract) and sourceTXID as signature in Vin
	tx := Transaction{
		nil,
		prepareInputLists(utxos, contractPubKeyHash, sourceTXID),
		prepareOutputLists(contractAddr, toAddr, amount, change, ""),
		tip,
		gasPrice,
		gasLimit,
	}
	tx.ID = tx.Hash()

	return tx, nil
}

func NewTransactionByVin(vinTxId []byte, vinVout int, vinPubkey []byte, voutValue uint64, voutPubKeyHash PubKeyHash, tip uint64) Transaction {
	tx := Transaction{
		ID: nil,
		Vin: []TXInput{
			{vinTxId, vinVout, nil, vinPubkey},
		},
		Vout: []TXOutput{
			{common.NewAmount(voutValue), voutPubKeyHash, ""},
		},
		Tip: common.NewAmount(tip),
	}
	tx.ID = tx.Hash()
	return tx
}

// NewGasRewardTx returns a reward to miner, earned for contract execution gas fee
func NewGasRewardTx(to Address, blockHeight uint64, actualGasCount *common.Amount, gasPrice *common.Amount) (Transaction, error) {
	fee := actualGasCount.Mul(gasPrice)
	bh := make([]byte, 8)
	binary.BigEndian.PutUint64(bh, uint64(blockHeight))

	txin := TXInput{nil, -1, bh, []byte("")}
	txout := NewTXOutput(fee, to)
	logger.WithFields(logger.Fields{
		"fee": fee,
	}).Info("NewGasRewardTx.")
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0)}
	tx.ID = tx.Hash()
	return tx, nil
}

// NewGasChangeTx returns a change to contract invoker, pay for the change of unused gas
func NewGasChangeTx(to Address, blockHeight uint64, actualGasCount *common.Amount, gasLimit *common.Amount, gasPrice *common.Amount) (Transaction, error) {
	if gasLimit.Cmp(actualGasCount) <= 0 {
		return Transaction{}, ErrNoGasChange
	}
	change, err := gasLimit.Sub(actualGasCount)

	if err != nil {
		return Transaction{}, err
	}
	changeValue := change.Mul(gasPrice)
	bh := make([]byte, 8)
	binary.BigEndian.PutUint64(bh, uint64(blockHeight))

	txin := TXInput{nil, -1, bh, []byte("")}
	txout := NewTXOutput(changeValue, to)
	logger.WithFields(logger.Fields{
		"change":      change.Uint64(),
		"changeValue": changeValue,
	}).Info("NewGasChangeTx gas changed.")
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0)}
	tx.ID = tx.Hash()
	return tx, nil
}

//GetContractAddress gets the smart contract's address if a transaction deploys a smart contract
func (tx *Transaction) GetContractAddress() Address {
	ctx := tx.ToContractTx()
	if ctx == nil {
		return NewAddress("")
	}

	return ctx.GetContractPubKeyHash().GenerateAddress()
}

//GetContract returns the smart contract code in a transaction
func (ctx *ContractTx) GetContract() string {
	return ctx.Vout[ContractTxouputIndex].Contract
}

//GetContractPubKeyHash returns the smart contract pubkeyhash in a transaction
func (ctx *ContractTx) GetContractPubKeyHash() PubKeyHash {
	return ctx.Vout[ContractTxouputIndex].PubKeyHash
}

//Execute executes the smart contract the transaction points to. it doesnt do anything if is a normal transaction
func (ctx *ContractTx) Execute(index UTXOIndex,
	scStorage *ScState,
	rewards map[string]string,
	engine ScEngine,
	currblkHeight uint64,
	parentBlk *Block) (uint64, []*Transaction, error) {

	if engine == nil {
		return 0, nil, nil
	}

	vout := ctx.Vout[ContractTxouputIndex]

	utxos := index.GetAllUTXOsByPubKeyHash([]byte(vout.PubKeyHash))
	//the smart contract utxo is always stored at index 0. If there is no utxos found, that means this transaction
	//is a smart contract deployment transaction, not a smart contract execution transaction.
	if utxos.Size() == 0 {
		return 0, nil, nil
	}

	prevUtxos, err := ctx.FindAllTxinsInUtxoPool(index)
	if err != nil {
		logger.WithError(err).WithFields(logger.Fields{
			"txid": hex.EncodeToString(ctx.ID),
		}).Warn("Transaction: cannot find vin while executing smart contract")
		return 0, nil, ErrUTXONotFound
	}

	function, args := util.DecodeScInput(vout.Contract)
	if function == "" {
		return 0, nil, ErrUnsupportedSourceType
	}

	totalArgs := util.PrepareArgs(args)
	address := vout.PubKeyHash.GenerateAddress()
	logger.WithFields(logger.Fields{
		"contract_address": address.String(),
		"invoked_function": function,
		"arguments":        totalArgs,
	}).Debug("Transaction: is executing the smart contract...")

	createContractUtxo, invokeUtxos := index.SplitContractUtxo([]byte(vout.PubKeyHash))
	if err := engine.SetExecutionLimits(ctx.GasLimit.Uint64(), DefaultLimitsOfTotalMemorySize); err != nil {
		logger.Info("Transaction: Execute SetExecutionLimits...")
		return 0, nil, ErrInvalidGasLimit
	}
	engine.ImportSourceCode(createContractUtxo.Contract)
	engine.ImportLocalStorage(scStorage)
	engine.ImportContractAddr(address)
	engine.ImportUTXOs(invokeUtxos)
	engine.ImportSourceTXID(ctx.ID)
	engine.ImportRewardStorage(rewards)
	engine.ImportTransaction(&ctx.Transaction)
	engine.ImportPrevUtxos(prevUtxos)
	engine.ImportCurrBlockHeight(currblkHeight)
	engine.ImportSeed(parentBlk.GetTimestamp())
	logger.Info("before engine.Execute...")
	_, err = engine.Execute(function, totalArgs)
	gasCount := engine.ExecutionInstructions()
	// record base gas
	baseGas, _ := ctx.GasCountOfTxBase()
	gasCount += baseGas.Uint64()
	if err != nil {
		return gasCount, nil, err
	}
	return gasCount, engine.GetGeneratedTXs(), err
}

//FindAllTxinsInUtxoPool Find the transaction in a utxo pool. Returns true only if all Vins are found in the utxo pool
func (tx *Transaction) FindAllTxinsInUtxoPool(utxoPool UTXOIndex) ([]*UTXO, error) {
	var res []*UTXO
	for _, vin := range tx.Vin {
		pubKeyHash, err := NewUserPubKeyHash(vin.PubKey)
		if err != nil {
			return nil, ErrNewUserPubKeyHash
		}
		utxo := utxoPool.FindUTXOByVin([]byte(pubKeyHash), vin.Txid, vin.Vout)
		if utxo == nil {
			logger.WithFields(logger.Fields{
				"txid":      hex.EncodeToString(tx.ID),
				"vin_id":    hex.EncodeToString(vin.Txid),
				"vin_index": vin.Vout,
			}).Warn("Transaction: Can not find vin")
			return nil, ErrTXInputNotFound
		}
		res = append(res, utxo)
	}
	return res, nil
}

func (tx *Transaction) IsIdentical(utxoIndex *UTXOIndex, tx2 *Transaction) bool {
	if tx2 == nil {
		return false
	}

	sender, recipient, amount, tip, err := tx.Describe(utxoIndex)
	if err != nil {
		return false
	}

	sender2, recipient2, amount2, tip2, err := tx2.Describe(utxoIndex)
	if err != nil {
		return false
	}

	if sender.String() != sender2.String() {
		return false
	}

	if recipient.String() != recipient2.String() {
		return false
	}

	if amount.Cmp(amount2) != 0 {
		return false
	}

	if tip.Cmp(tip2) != 0 {
		return false
	}

	return true

}

func (tx *Transaction) MatchRewards(rewardStorage map[string]string) bool {

	if tx == nil {
		logger.Debug("Transaction: does not exist")
		return false
	}

	if !tx.IsRewardTx() {
		logger.Debug("Transaction: is not a reward transaction")
		return false
	}

	for _, vout := range tx.Vout {
		if !vout.IsFoundInRewardStorage(rewardStorage) {
			return false
		}
	}
	return len(rewardStorage) == len(tx.Vout)
}

// String returns a human-readable representation of a transaction
func (tx *Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("\n--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", []byte(output.PubKeyHash)))
		lines = append(lines, fmt.Sprintf("       Contract: %s", output.Contract))
	}
	lines = append(lines, "\n")

	return strings.Join(lines, "\n")
}

//calculateChange calculates the change
func calculateChange(input, amount, tip *common.Amount, gasLimit *common.Amount, gasPrice *common.Amount) (*common.Amount, error) {
	change, err := input.Sub(amount)
	if err != nil {
		return nil, ErrInsufficientFund
	}

	change, err = change.Sub(tip)
	if err != nil {
		return nil, ErrInsufficientFund
	}
	change, err = change.Sub(gasLimit.Mul(gasPrice))
	if err != nil {
		return nil, ErrInsufficientFund
	}
	return change, nil
}

//prepareInputLists prepares a list of txinputs for a new transaction
func prepareInputLists(utxos []*UTXO, publicKey []byte, signature []byte) []TXInput {
	var inputs []TXInput

	// Build a list of inputs
	for _, utxo := range utxos {
		input := TXInput{utxo.Txid, utxo.TxIndex, signature, publicKey}
		inputs = append(inputs, input)
	}

	return inputs
}

//calculateUtxoSum calculates the total amount of all input utxos
func calculateUtxoSum(utxos []*UTXO) *common.Amount {
	sum := common.NewAmount(0)
	for _, utxo := range utxos {
		sum = sum.Add(utxo.Value)
	}
	return sum
}

//preapreOutPutLists prepares a list of txoutputs for a new transaction
func prepareOutputLists(from, to Address, amount *common.Amount, change *common.Amount, contract string) []TXOutput {

	var outputs []TXOutput
	toAddr := to

	if toAddr.String() == "" {
		toAddr = NewContractPubKeyHash().GenerateAddress()
	}

	if contract != "" {
		outputs = append(outputs, *NewContractTXOutput(toAddr, contract))
	}

	outputs = append(outputs, *NewTXOutput(amount, toAddr))
	logger.WithFields(logger.Fields{
		"change": change,
	}).Warn("prepareOutputLists")
	if !change.IsZero() {
		logger.WithFields(logger.Fields{
			"change": change,
		}).Warn("prepareOutputLists not zero")
		outputs = append(outputs, *NewTXOutput(change, from))
	}
	return outputs
}

func (tx *Transaction) ToProto() proto.Message {

	var vinArray []*corepb.TXInput
	for _, txin := range tx.Vin {
		vinArray = append(vinArray, txin.ToProto().(*corepb.TXInput))
	}

	var voutArray []*corepb.TXOutput
	for _, txout := range tx.Vout {
		voutArray = append(voutArray, txout.ToProto().(*corepb.TXOutput))
	}
	if tx.GasLimit == nil {
		tx.GasLimit = common.NewAmount(0)
	}
	if tx.GasPrice == nil {
		tx.GasPrice = common.NewAmount(0)
	}
	return &corepb.Transaction{
		Id:       tx.ID,
		Vin:      vinArray,
		Vout:     voutArray,
		Tip:      tx.Tip.Bytes(),
		GasLimit: tx.GasLimit.Bytes(),
		GasPrice: tx.GasPrice.Bytes(),
	}
}

func (tx *Transaction) FromProto(pb proto.Message) {
	tx.ID = pb.(*corepb.Transaction).GetId()
	tx.Tip = common.NewAmountFromBytes(pb.(*corepb.Transaction).GetTip())

	var vinArray []TXInput
	txin := TXInput{}
	for _, txinpb := range pb.(*corepb.Transaction).GetVin() {
		txin.FromProto(txinpb)
		vinArray = append(vinArray, txin)
	}
	tx.Vin = vinArray

	var voutArray []TXOutput
	txout := TXOutput{}
	for _, txoutpb := range pb.(*corepb.Transaction).GetVout() {
		txout.FromProto(txoutpb)
		voutArray = append(voutArray, txout)
	}
	tx.Vout = voutArray

	tx.GasLimit = common.NewAmountFromBytes(pb.(*corepb.Transaction).GetGasLimit())
	tx.GasPrice = common.NewAmountFromBytes(pb.(*corepb.Transaction).GetGasPrice())
}

func (tx *Transaction) GetSize() int {
	rawBytes, err := proto.Marshal(tx.ToProto())
	if err != nil {
		logger.Warn("Transaction: Transaction can not be marshalled!")
		return 0
	}
	return len(rawBytes)
}

// GasCountOfTxBase calculate the actual amount for a tx with data
func (ctx *ContractTx) GasCountOfTxBase() (*common.Amount, error) {
	txGas := MinGasCountPerTransaction
	if dataLen := ctx.DataLen(); dataLen > 0 {
		dataGas := common.NewAmount(uint64(dataLen)).Mul(GasCountPerByte)
		baseGas := txGas.Add(dataGas)
		txGas = baseGas
	}
	return txGas, nil
}

// DataLen return the length of payload
func (ctx *ContractTx) DataLen() int {
	return len([]byte(ctx.GetContract()))
}

// GetDefaultFromPubKeyHash returns the first from address public key hash
func (tx *Transaction) GetDefaultFromPubKeyHash() PubKeyHash {
	if tx.Vin == nil || len(tx.Vin) <= 0 {
		return nil
	}
	vin := tx.Vin[0]
	pubKeyHash, err := NewUserPubKeyHash(vin.PubKey)
	if err != nil {
		return nil
	}
	return pubKeyHash
}
