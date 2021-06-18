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
	utxopb "github.com/dappley/go-dappley/core/utxo/pb"
	"github.com/dappley/go-dappley/logic/lutxo"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/dappley/go-dappley/core/scState"
	"github.com/dappley/go-dappley/core/transaction"
	"github.com/dappley/go-dappley/core/transaction/pb"
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
	bm             *lblockchain.BlockchainManager
	node           *network.Node
	dynasty        *consensus.Dynasty
	utxoIndex      *lutxo.UTXOIndex //rpc cache
	dbUtxoIndex    *lutxo.UTXOIndex
	blockMaxHeight uint64
	mutex          sync.Mutex
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
	addressAccount := account.NewTransactionAccountByAddress(account.NewAddress(address))
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

func (rpcService *RpcService) RpcGetUTXOsAccordingToAmount(server rpcpb.RpcService_RpcGetUTXOsAccordingToAmountServer) error {
	bc := rpcService.GetBlockchain()
	rpcService.mutex.Lock()
	if rpcService.utxoIndex == nil || rpcService.blockMaxHeight < bc.GetMaxHeight() {
		errFlag := true
		if rpcService.utxoIndex, errFlag = bc.GetUpdatedUTXOIndex(); !errFlag {
			logger.Warn("RpcSendTransaction update utxoIndex error")
		}
		rpcService.blockMaxHeight = bc.GetMaxHeight()
	}
	rpcService.mutex.Unlock()

	req, err := server.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	amount := common.NewAmount(uint64(req.Amount))
	acc := account.NewTransactionAccountByAddress(account.NewAddress(req.Address))
	if !acc.IsValid() {
		return status.Error(codes.InvalidArgument, logic.ErrInvalidAddress.Error())
	}
	response := rpcpb.GetUTXOResponse{}
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
	utxos,err := rpcService.utxoIndex.GetUTXOsAccordingToAmount([]byte(acc.GetPubKeyHash()),amount)
	logger.Info("rpcService utxoIndex GetUTXOsAccordingToAmount len utxos:",len(utxos)," pubkeyHash:",acc.GetPubKeyHash()," error:",err)
	if len(utxos) == 0 {
		err := server.Send(&response)
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("Server Send Failed!")
			return err
		}
	} else {
		count := 0
		for i:=0;i<len(utxos);i++ {
			count++
			response.Utxos = append(response.Utxos, utxos[i].ToProto().(*utxopb.Utxo))
			if count%1000 == 0 || count == len(utxos) {
				err := server.Send(&response)
				if err != nil {
					logger.WithFields(logger.Fields{
						"error": err,
					}).Error("Server Send Failed!")
					return err
				}
				if count != len(utxos) {
					_, err = server.Recv()
					if err == io.EOF {
						return nil
					}
					if err != nil {
						return err
					}
					response = rpcpb.GetUTXOResponse{}
				}
			}
		}
	}
	return nil
}

func (rpcService *RpcService) RpcGetUTXO(server rpcpb.RpcService_RpcGetUTXOServer) error {
	bc := rpcService.GetBlockchain()
	rpcService.mutex.Lock()
	if rpcService.dbUtxoIndex == nil || rpcService.blockMaxHeight < bc.GetMaxHeight() {
		rpcService.dbUtxoIndex = lutxo.NewUTXOIndex(bc.GetUtxoCache())
		rpcService.blockMaxHeight = bc.GetMaxHeight()
	}
	rpcService.mutex.Unlock()

	req, err := server.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	acc := account.NewTransactionAccountByAddress(account.NewAddress(req.Address))
	if !acc.IsValid() {
		return status.Error(codes.InvalidArgument, logic.ErrInvalidAddress.Error())
	}
	response := rpcpb.GetUTXOResponse{}
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
	utxos := rpcService.dbUtxoIndex.GetAllUTXOsByPubKeyHash(acc.GetPubKeyHash())
	if len(utxos.Indices) == 0 {
		err := server.Send(&response)
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("Server Send Failed!")
			return err
		}
	} else {
		count := 0
		for _, utxo := range utxos.Indices {
			count++
			response.Utxos = append(response.Utxos, utxo.ToProto().(*utxopb.Utxo))
			if count%1000 == 0 || count == len(utxos.Indices) {
				err := server.Send(&response)
				if err != nil {
					logger.WithFields(logger.Fields{
						"error": err,
					}).Error("Server Send Failed!")
					return err
				}
				if count != len(utxos.Indices) {
					_, err = server.Recv()
					if err == io.EOF {
						return nil
					}
					if err != nil {
						return err
					}
					response = rpcpb.GetUTXOResponse{}
				}
			}
		}
	}
	return nil
}

func (rpcService *RpcService) RpcGetBlocks(ctx context.Context, in *rpcpb.GetBlocksRequest) (*rpcpb.GetBlocksResponse, error) {
	result := &rpcpb.GetBlocksResponse{}
	blk := rpcService.findBlockInRequestHash(in.GetStartBlockHashes())
	if blk.GetTimestamp() == -1 {
		result.Blocks = append(result.Blocks, blk.ToProto().(*blockpb.Block))
		return result, nil
	}
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
	if len(startBlockHashes) > 0 {
		blk.SetTimestamp(-1)
	}
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

	adaptedTx := transaction.NewTxAdapter(tx)
	if !adaptedTx.IsNormal() && !adaptedTx.IsContract() {
		return nil, status.Error(codes.InvalidArgument, "transaction type error, must be normal or contract")
	}

	if adaptedTx.IsContract() && adaptedTx.GasPrice.Cmp(common.NewAmount(0)) < 0 {
		return nil, status.Error(codes.InvalidArgument, "gas price error, must be a positive number")
	}

	bc := rpcService.GetBlockchain()
	rpcService.mutex.Lock()
	if rpcService.utxoIndex == nil || rpcService.blockMaxHeight < bc.GetMaxHeight() {
		errFlag := true
		if rpcService.utxoIndex, errFlag = bc.GetUpdatedUTXOIndex(); !errFlag {
			logger.Warn("RpcSendTransaction update utxoIndex error")
		}
		rpcService.blockMaxHeight = bc.GetMaxHeight()
	}
	rpcService.mutex.Unlock()

	if err := ltransaction.VerifyTransaction(rpcService.utxoIndex, tx, 0); err != nil {
		logger.Warn(err.Error())
		return nil, status.Error(codes.FailedPrecondition, lblockchain.ErrTransactionVerifyFailed.Error())
	}

	engine := vm.NewV8Engine()
	defer engine.DestroyEngine()

	if err := ltransaction.CheckContractSyntaxTransaction(engine, tx); err != nil {
		logger.WithFields(logger.Fields{
			"error": err,
		}).Error("Smart Contract Deployed Failed!")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rpcService.mutex.Lock()
	if !rpcService.utxoIndex.UpdateUtxo(tx) {
		rpcService.mutex.Unlock()
		logger.Error("updateUTXO failed.")
		return nil, status.Error(codes.InvalidArgument, "updateUTXO failed")
	}
	bc.GetTxPool().Push(*tx)
	rpcService.mutex.Unlock()
	bc.GetTxPool().BroadcastTx(tx)

	var generatedContractAddress = ""
	if adaptedTx.IsContract() {
		ctx := ltransaction.NewTxContract(tx)
		contractAddr := ctx.GetContractAddress()
		generatedContractAddress = contractAddr.String()
		logger.WithFields(logger.Fields{
			"Contract Address": generatedContractAddress,
		}).Info("Smart Contract has been received.")
	}

	return &rpcpb.SendTransactionResponse{GeneratedContractAddress: generatedContractAddress}, nil
}

// RpcSendBatchTransaction sends a batch of ordered transactions to blockchain created by account
func (rpcService *RpcService) RpcSendBatchTransaction(ctx context.Context, in *rpcpb.SendBatchTransactionRequest) (*rpcpb.SendBatchTransactionResponse, error) {
	var respon []proto.Message
	utxoIndex, _ := rpcService.GetBlockchain().GetUpdatedUTXOIndex()

	txs := []transaction.Transaction{}
	for _, txInReq := range in.Transactions {
		tx := transaction.Transaction{nil, nil, nil, common.NewAmount(0), common.NewAmount(0), common.NewAmount(0), time.Now().UnixNano() / 1e6, transaction.TxTypeDefault}
		tx.FromProto(txInReq)
		txs = append(txs, tx)
	}

	// verify transactions
	verifiedTxs := []transaction.Transaction{}
	st := status.New(codes.OK, "")
	for _, tx := range txs {
		adaptedTx := transaction.NewTxAdapter(&tx)
		if !adaptedTx.IsNormal() && !adaptedTx.IsContract() {
			st = status.New(codes.Unknown, "one or more transactions are invalid")
			respon = append(respon, &rpcpb.SendTransactionStatus{
				Txid:    tx.ID,
				Code:    uint32(codes.InvalidArgument),
				Message: "transaction type error, must be normal or contract",
			})
			continue
		}

		if adaptedTx.IsContract() && adaptedTx.GasPrice.Cmp(common.NewAmount(0)) <= 0 {
			st = status.New(codes.Unknown, "one or more transactions are invalid")
			respon = append(respon, &rpcpb.SendTransactionStatus{
				Txid:    tx.ID,
				Code:    uint32(codes.InvalidArgument),
				Message: "gas price error, must be a positive number",
			})
			continue
		}

		if err := ltransaction.VerifyTransaction(utxoIndex, &tx, 0); err != nil {
			st = status.New(codes.Unknown, "one or more transactions are invalid")
			// add invalid transactions to response details if exists
			respon = append(respon, &rpcpb.SendTransactionStatus{
				Txid:    tx.ID,
				Code:    uint32(codes.FailedPrecondition),
				Message: lblockchain.ErrTransactionVerifyFailed.Error(),
			})
			continue
		}

		utxoIndex.UpdateUtxo(&tx)
		rpcService.GetBlockchain().GetTxPool().Push(tx)
		verifiedTxs = append(verifiedTxs, tx)

		respon = append(respon, &rpcpb.SendTransactionStatus{
			Txid:    tx.ID,
			Code:    uint32(codes.OK),
			Message: "",
		})
	}
	rpcService.GetBlockchain().GetTxPool().BroadcastBatchTxs(verifiedTxs)

	st, _ = st.WithDetails(respon...)
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
	bc := rpcService.GetBlockchain()
	utxoIndex := lutxo.NewUTXOIndex(bc.GetUtxoCache())
	txs := bc.GetTxPool().GetAllTransactions(utxoIndex)
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

	contractTx := ltransaction.NewTxContract(tx)
	if contractTx == nil {
		return nil, status.Error(codes.FailedPrecondition, "cannot estimate normal transaction")
	}
	utxoIndex, errFlag := rpcService.GetBlockchain().GetUpdatedUTXOIndex()
	if !errFlag {
		logger.Warn("RpcEstimateGase error")
	}
	if err := contractTx.VerifyInEstimate(utxoIndex); err != nil {
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

	if contractAddr == "" || queryKey == "" {
		return nil, status.Error(codes.InvalidArgument, "contract query params error")
	}
	scState := scState.NewScState(rpcService.GetBlockchain().GetUtxoCache())
	resultValue := scState.GetStateValue(contractAddr, queryKey)

	return &rpcpb.ContractQueryResponse{Key: queryKey, Value: resultValue}, nil
}
