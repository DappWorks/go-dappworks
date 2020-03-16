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
package rpc

import (
	"context"
	"github.com/dappley/go-dappley/consensus"
	"strconv"
	"strings"
	"time"

	"github.com/dappley/go-dappley/core/scState"
	"github.com/dappley/go-dappley/core/transaction"
	"github.com/dappley/go-dappley/core/transaction/pb"
	"github.com/dappley/go-dappley/core/utxo/pb"
	"github.com/dappley/go-dappley/logic/ltransaction"
	"github.com/dappley/go-dappley/logic/transactionpool"

	"github.com/dappley/go-dappley/core/block"
	"github.com/dappley/go-dappley/core/block/pb"
	"github.com/dappley/go-dappley/logic/lblockchain"

	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/crypto/byteutils"

	"github.com/golang/protobuf/proto"
	logger "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/logic"
	"github.com/dappley/go-dappley/network"
	"github.com/dappley/go-dappley/rpc/pb"
	"github.com/dappley/go-dappley/vm"
)

const (
	ProtoVersion                   = "1.0.0"
	MaxGetBlocksCount       int32  = 500
	MinUtxoBlockHeaderCount uint64 = 6
)

type RpcService struct {
	bm      *lblockchain.BlockchainManager
	node    *network.Node
	dynasty *consensus.Dynasty
}

func (rpcSerivce *RpcService) GetBlockchain() *lblockchain.Blockchain {
	if rpcSerivce.bm == nil {
		return nil
	}

	return rpcSerivce.bm.Getblockchain()
}

func (rpcService *RpcService) RpcGetVersion(ctx context.Context, in *rpcpb.GetVersionRequest) (*rpcpb.GetVersionResponse, error) {
	clientProtoVersions := strings.Split(in.GetProtoVersion(), ".")

	if len(clientProtoVersions) != 3 {
		return nil, status.Error(codes.InvalidArgument, "proto version not supported")
	}

	serverProtoVersions := strings.Split(ProtoVersion, ".")

	// Major version must equal
	if serverProtoVersions[0] != clientProtoVersions[0] {
		return nil, status.Error(codes.Unimplemented, "major version mismatch")
	}

	return &rpcpb.GetVersionResponse{ProtoVersion: ProtoVersion, ServerVersion: ""}, nil
}

func (rpcService *RpcService) RpcGetBalance(ctx context.Context, in *rpcpb.GetBalanceRequest) (*rpcpb.GetBalanceResponse, error) {
	address := in.GetAddress()
	addressAccount := account.NewContractAccountByAddress(account.NewAddress(address))
	if !addressAccount.IsValid() {
		return nil, status.Error(codes.InvalidArgument, account.ErrInvalidAddress.Error())
	}

	amount, err := logic.GetBalance(addressAccount.GetAddress(), rpcService.GetBlockchain())
	if err != nil {
		switch err {
		case logic.ErrInvalidAddress:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}
	return &rpcpb.GetBalanceResponse{Amount: amount.Int64()}, nil
}

func (rpcService *RpcService) RpcGetBlockchainInfo(ctx context.Context, in *rpcpb.GetBlockchainInfoRequest) (*rpcpb.GetBlockchainInfoResponse, error) {
	tailBlock, err := rpcService.GetBlockchain().GetTailBlock()
	if err != nil {
		switch err {
		case lblockchain.ErrBlockDoesNotExist:
			return nil, status.Error(codes.Internal, err.Error())
		default:
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return &rpcpb.GetBlockchainInfoResponse{
		TailBlockHash: rpcService.GetBlockchain().GetTailBlockHash(),
		BlockHeight:   rpcService.GetBlockchain().GetMaxHeight(),
		Producers:     rpcService.dynasty.GetProducers(),
		Timestamp:     tailBlock.GetTimestamp(),
	}, nil
}

func (rpcService *RpcService) RpcGetUTXO(ctx context.Context, in *rpcpb.GetUTXORequest) (*rpcpb.GetUTXOResponse, error) {
	utxoIndex := rpcService.GetBlockchain().GetUpdatedUTXOIndex()

	acc := account.NewContractAccountByAddress(account.NewAddress(in.GetAddress()))

	if !acc.IsValid() {
		return nil, status.Error(codes.InvalidArgument, logic.ErrInvalidAddress.Error())
	}

	utxos := utxoIndex.GetAllUTXOsByPubKeyHash(acc.GetPubKeyHash())
	response := rpcpb.GetUTXOResponse{}
	for _, utxo := range utxos.Indices {
		response.Utxos = append(response.Utxos, utxo.ToProto().(*utxopb.Utxo))
	}

	//TODO Race condition Blockchain update after GetUTXO
	getHeaderCount := MinUtxoBlockHeaderCount
	if int(getHeaderCount) < len(rpcService.dynasty.GetProducers()) {
		getHeaderCount = uint64(len(rpcService.dynasty.GetProducers()))
	}

	tailHeight := rpcService.GetBlockchain().GetMaxHeight()
	if getHeaderCount > tailHeight {
		getHeaderCount = tailHeight
	}

	for i := uint64(0); i < getHeaderCount; i++ {
		blk, err := rpcService.GetBlockchain().GetBlockByHeight(tailHeight - uint64(i))
		if err != nil {
			break
		}

		response.BlockHeaders = append(response.BlockHeaders, blk.GetHeader().ToProto().(*blockpb.BlockHeader))
	}

	return &response, nil
}

// RpcGetBlocks Get blocks in blockchain from head to tail
func (rpcService *RpcService) RpcGetBlocks(ctx context.Context, in *rpcpb.GetBlocksRequest) (*rpcpb.GetBlocksResponse, error) {
	blk := rpcService.findBlockInRequestHash(in.GetStartBlockHashes())

	// Reach the blockchain's tail
	if blk.GetHeight() >= rpcService.GetBlockchain().GetMaxHeight() {
		return &rpcpb.GetBlocksResponse{}, nil
	}

	var blocks []*block.Block
	maxBlockCount := in.GetMaxCount()
	if maxBlockCount > MaxGetBlocksCount {
		return nil, status.Error(codes.InvalidArgument, "block count overflow")
	}

	blk, err := rpcService.GetBlockchain().GetBlockByHeight(blk.GetHeight() + 1)
	for i := int32(0); i < maxBlockCount && err == nil; i++ {
		blocks = append(blocks, blk)
		blk, err = rpcService.GetBlockchain().GetBlockByHeight(blk.GetHeight() + 1)
	}

	result := &rpcpb.GetBlocksResponse{}

	for _, blk = range blocks {
		result.Blocks = append(result.Blocks, blk.ToProto().(*blockpb.Block))
	}

	return result, nil
}

func (rpcService *RpcService) findBlockInRequestHash(startBlockHashes [][]byte) *block.Block {
	for _, hash := range startBlockHashes {
		// hash in blockchain, return
		if blk, err := rpcService.GetBlockchain().GetBlockByHash(hash); err == nil {
			return blk
		}
	}

	// Return Genesis Block
	blk, _ := rpcService.GetBlockchain().GetBlockByHeight(0)
	return blk
}

// RpcGetBlockByHash Get single block in blockchain by hash
func (rpcService *RpcService) RpcGetBlockByHash(ctx context.Context, in *rpcpb.GetBlockByHashRequest) (*rpcpb.GetBlockByHashResponse, error) {
	blk, err := rpcService.GetBlockchain().GetBlockByHash(in.GetHash())

	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &rpcpb.GetBlockByHashResponse{Block: blk.ToProto().(*blockpb.Block)}, nil
}

// RpcGetBlockByHeight Get single block in blockchain by height
func (rpcService *RpcService) RpcGetBlockByHeight(ctx context.Context, in *rpcpb.GetBlockByHeightRequest) (*rpcpb.GetBlockByHeightResponse, error) {
	blk, err := rpcService.GetBlockchain().GetBlockByHeight(in.GetHeight())

	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &rpcpb.GetBlockByHeightResponse{Block: blk.ToProto().(*blockpb.Block)}, nil
}

// RpcSendTransaction Send transaction to blockchain created by account account
func (rpcService *RpcService) RpcSendTransaction(ctx context.Context, in *rpcpb.SendTransactionRequest) (*rpcpb.SendTransactionResponse, error) {

	tx := &transaction.Transaction{nil, nil, nil, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), time.Now().UnixNano() / 1e6, transaction.TxTypeDefault}

	tx.FromProto(in.GetTransaction())

	if !tx.IsNormal() && !tx.IsContract() {
		return nil, status.Error(codes.InvalidArgument, "transaction type error, must be normal or contract")
	}

	utxoIndex := rpcService.GetBlockchain().GetUpdatedUTXOIndex()

	if err := ltransaction.VerifyTransaction(utxoIndex, tx, 0); err != nil {
		logger.Warn(err.Error())
		return nil, status.Error(codes.FailedPrecondition, lblockchain.ErrTransactionVerifyFailed.Error())
	}

	engine := vm.NewV8Engine()
	defer engine.DestroyEngine()
	err := ltransaction.CheckContractSyntaxTransaction(engine, tx)

	if err != nil {
		logger.WithFields(logger.Fields{
			"error": err,
		}).Error("Smart Contract Deployed Failed!")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rpcService.GetBlockchain().GetTxPool().Push(*tx)
	rpcService.GetBlockchain().GetTxPool().BroadcastTx(tx)

	var generatedContractAddress = ""
	if tx.IsContract() {
		ctx := ltransaction.NewTxContract(tx)
		contractAddr := ctx.GetContractAddress()
		generatedContractAddress = contractAddr.String()
		logger.WithFields(logger.Fields{
			"contractAddr": generatedContractAddress,
		}).Info("Smart Contract Deployed Successful!")
	}

	return &rpcpb.SendTransactionResponse{GeneratedContractAddress: generatedContractAddress}, nil
}

// RpcSendBatchTransaction sends a batch of transactions to blockchain created by account account
func (rpcService *RpcService) RpcSendBatchTransaction(ctx context.Context, in *rpcpb.SendBatchTransactionRequest) (*rpcpb.SendBatchTransactionResponse, error) {
	statusCode := codes.OK
	var details []proto.Message
	utxoIndex := rpcService.GetBlockchain().GetUpdatedUTXOIndex()

	txMap := make(map[int]transaction.Transaction, len(in.Transactions))
	txs := []transaction.Transaction{}
	for key, txInReq := range in.Transactions {
		tx := transaction.Transaction{nil, nil, nil, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), time.Now().UnixNano() / 1e6, transaction.TxTypeDefault}

		tx.FromProto(txInReq)
		txs = append(txs, tx)
		txMap[key] = tx
	}

	// verify dependent transactions within batch of transactions
	lastTxsLen := 0
	verifiedTxs := []transaction.Transaction{}
	for len(txMap) != lastTxsLen {
		lastTxsLen = len(txMap)
		for key, tx := range txs {
			if _, ok := txMap[key]; !ok {
				continue
			}

			if !tx.IsNormal() && !tx.IsContract() {
				if statusCode == codes.OK {
					statusCode = codes.Unknown
				}
				details = append(details, &rpcpb.SendTransactionStatus{
					Txid:    tx.ID,
					Code:    uint32(codes.InvalidArgument),
					Message: "transaction type error, must be normal or contract",
				})
				delete(txMap, key)
				continue
			}

			if err := ltransaction.VerifyTransaction(utxoIndex, &tx, 0); err != nil {
				continue
			}

			utxoIndex.UpdateUtxo(&tx)
			rpcService.GetBlockchain().GetTxPool().Push(tx)
			verifiedTxs = append(verifiedTxs, tx)

			details = append(details, &rpcpb.SendTransactionStatus{
				Txid:    tx.ID,
				Code:    uint32(codes.OK),
				Message: "",
			})
			delete(txMap, key)
		}
	}

	rpcService.GetBlockchain().GetTxPool().BroadcastBatchTxs(verifiedTxs)

	st := status.New(codes.OK, "")
	// add invalid transactions to response details if exists
	if statusCode == codes.Unknown || len(txMap) > 0 {
		st = status.New(codes.Unknown, "one or more transactions are invalid")
		for _, tx := range txMap {
			details = append(details, &rpcpb.SendTransactionStatus{
				Txid:    tx.ID,
				Code:    uint32(codes.FailedPrecondition),
				Message: lblockchain.ErrTransactionVerifyFailed.Error(),
			})

		}
	}
	st, _ = st.WithDetails(details...)

	return &rpcpb.SendBatchTransactionResponse{}, st.Err()
}

func (rpcService *RpcService) RpcGetNewTransaction(in *rpcpb.GetNewTransactionRequest, stream rpcpb.RpcService_RpcGetNewTransactionServer) error {
	var txHandler interface{}

	quitCh := make(chan bool, 1)

	txHandler = func(tx *transaction.Transaction) {
		response := &rpcpb.GetNewTransactionResponse{Transaction: tx.ToProto().(*transactionpb.Transaction)}
		err := stream.Send(response)
		if err != nil {
			logger.WithError(err).Info("RPCService: failed to send transaction to account.")
			rpcService.GetBlockchain().GetTxPool().EventBus.Unsubscribe(transactionpool.NewTransactionTopic, txHandler)
			quitCh <- true
		}
	}

	rpcService.GetBlockchain().GetTxPool().EventBus.SubscribeAsync(transactionpool.NewTransactionTopic, txHandler, false)
	<-quitCh
	return nil
}

func (rpcService *RpcService) RpcSubscribe(in *rpcpb.SubscribeRequest, stream rpcpb.RpcService_RpcSubscribeServer) error {
	quitCh := make(chan bool, 1)
	var cb interface{}
	cb = func(event *scState.Event) {
		response := &rpcpb.SubscribeResponse{Data: event.GetData()}
		err := stream.Send(response)
		if err != nil {
			logger.WithError(err).WithFields(logger.Fields{
				"topic": event.GetTopic(),
				"data":  event.GetData(),
			}).Info("RPCService: failed to send published data")
			rpcService.GetBlockchain().GetEventManager().Unsubscribe(event.GetTopic(), cb)
			quitCh <- true
		}
	}
	rpcService.GetBlockchain().GetEventManager().SubscribeMultiple(in.Topics, cb)
	<-quitCh
	return nil
}

func (rpcService *RpcService) IsPrivate() bool { return false }

// RpcGetAllTransactionsFromTxPool get all transactions from transactionpool
func (rpcService *RpcService) RpcGetAllTransactionsFromTxPool(ctx context.Context, in *rpcpb.GetAllTransactionsRequest) (*rpcpb.GetAllTransactionsResponse, error) {
	txs := rpcService.GetBlockchain().GetTxPool().GetTransactions()
	result := &rpcpb.GetAllTransactionsResponse{}
	for _, tx := range txs {
		result.Transactions = append(result.Transactions, tx.ToProto().(*transactionpb.Transaction))
	}
	return result, nil
}

// RpcGetLastIrreversibleBlock get last irreversible block
func (rpcService *RpcService) RpcGetLastIrreversibleBlock(ctx context.Context, in *rpcpb.GetLastIrreversibleBlockRequest) (*rpcpb.GetLastIrreversibleBlockResponse, error) {
	blk, err := rpcService.GetBlockchain().GetLIB()

	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &rpcpb.GetLastIrreversibleBlockResponse{Block: blk.ToProto().(*blockpb.Block)}, nil
}

// RpcEstimateGas estimate gas value of contract deploy and execution.
func (rpcService *RpcService) RpcEstimateGas(ctx context.Context, in *rpcpb.EstimateGasRequest) (*rpcpb.EstimateGasResponse, error) {

	tx := &transaction.Transaction{nil, nil, nil, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), time.Now().UnixNano() / 1e6, transaction.TxTypeDefault}

	tx.FromProto(in.GetTransaction())

	if !tx.IsContract() {
		return nil, status.Error(codes.InvalidArgument, "transaction type must be contract")
	}
	contractTx := ltransaction.NewTxContract(tx)
	if contractTx == nil {
		return nil, status.Error(codes.FailedPrecondition, "cannot estimate normal transaction")
	}
	utxoIndex := rpcService.GetBlockchain().GetUpdatedUTXOIndex()

	err := contractTx.VerifyInEstimate(utxoIndex)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	tx.GasLimit = common.NewAmount(vm.MaxLimitsOfExecutionInstructions)
	tailBlk, _ := rpcService.GetBlockchain().GetTailBlock()
	gasCount, err := vm.EstimateGas(
		tx,
		tailBlk,
		rpcService.GetBlockchain().GetUtxoCache(),
		rpcService.GetBlockchain().GetDb(),
	)
	return &rpcpb.EstimateGasResponse{GasCount: byteutils.FromUint64(gasCount)}, err
}

// RpcGasPrice returns current gas price.
func (rpcService *RpcService) RpcGasPrice(ctx context.Context, in *rpcpb.GasPriceRequest) (*rpcpb.GasPriceResponse, error) {
	gasPrice := rpcService.GetBlockchain().GasPrice()
	return &rpcpb.GasPriceResponse{GasPrice: byteutils.FromUint64(gasPrice)}, nil
}

// RpcContractQuery returns the query result of contract storage
func (rpcService *RpcService) RpcContractQuery(ctx context.Context, in *rpcpb.ContractQueryRequest) (*rpcpb.ContractQueryResponse, error) {

	contractAddr := in.ContractAddr
	queryKey := in.Key
	queryValue := in.Value

	if contractAddr == "" || (queryKey == "" && queryValue == "") {
		return nil, status.Error(codes.InvalidArgument, "contract query params error")
	}
	scState := scState.LoadScStateFromDatabase(rpcService.GetBlockchain().GetDb())
	var resultKey = ""
	var resultValue = ""
	// storage data has been JSON.stringfy before
	if queryKey != "" {
		resultKey = queryKey
		resultValue = scState.Get(contractAddr, queryKey)
		resultValue, _ = strconv.Unquote(resultValue)
	} else {
		resultValue = queryValue
		queryValue = strconv.Quote(queryValue)
		resultKey = scState.GetByValue(contractAddr, queryValue)
	}
	return &rpcpb.ContractQueryResponse{Key: resultKey, Value: resultValue}, nil
}
