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
	"errors"
	"fmt"
	"github.com/dappley/go-dappley/client"
	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/logic"
	"github.com/dappley/go-dappley/network"
	"github.com/dappley/go-dappley/network/pb"
	"github.com/dappley/go-dappley/rpc/pb"
	"github.com/dappley/go-dappley/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

const ProtoVersion = "1.0.0"

type RpcService struct {
	node *network.Node
}

func (rpcService *RpcService) RpcGetVersion(ctx context.Context, in *rpcpb.GetVersionRequest) (*rpcpb.GetVersionResponse, error) {
	clientProtoVersions := strings.Split(in.ProtoVersion, ".")

	if len(clientProtoVersions) != 3 {
		return &rpcpb.GetVersionResponse{ErrorCode: ProtoVersionNotSupport, ProtoVersion: ProtoVersion, ""}, nil
	}

	serverProtoVersions := strings.Split(ProtoVersion, ".")

	// Major version must equal
	if serverProtoVersions[0] != clientProtoVersions[0] {
		return &rpcpb.GetVersionResponse{ErrorCode: ProtoVersionNotSupport, ProtoVersion: ProtoVersion, ""}, nil
	}

	return &rpcpb.GetVersionResponse{ErrorCode: OK, ProtoVersion: ProtoVersion}, nil
}

// SayHello implements helloworld.GreeterServer
func (rpcSerivce *RpcService) RpcCreateWallet(ctx context.Context, in *rpcpb.CreateWalletRequest) (*rpcpb.CreateWalletResponse, error) {
	passPhrase := in.Passphrase
	fmt.Println(passPhrase)
	msg := ""
	if len(passPhrase) == 0 {
		logrus.Error("CreateWallet: Password is empty!")
		msg = "Create Wallet: Error"
		return &rpcpb.CreateWalletResponse{
			Message: msg,
			Address: ""}, nil
	}
	wallet, err := logic.CreateWalletWithpassphrase(passPhrase)
	if err != nil {
		msg = "Create Wallet: Error"
	}
	addr := wallet.GetAddress().Address
	fmt.Println(addr)
	msg = "Create Wallet: "
	return &rpcpb.CreateWalletResponse{
		Message: msg,
		Address: addr}, nil
}

func (rpcSerivce *RpcService) RpcGetBalance(ctx context.Context, in *rpcpb.GetBalanceRequest) (*rpcpb.GetBalanceResponse, error) {
	return &rpcpb.GetBalanceResponse{Message: "Hello " + in.Name}, nil
}

func (rpcSerivce *RpcService) RpcSend(ctx context.Context, in *rpcpb.SendRequest) (*rpcpb.SendResponse, error) {
	sendFromAddress := core.NewAddress(in.From)
	sendToAddress := core.NewAddress(in.To)
	sendAmount := common.NewAmountFromBytes(in.Amount)

	if sendAmount.Validate() != nil || sendAmount.IsZero() {
		return &rpcpb.SendResponse{Message: "Invalid send amount"}, core.ErrInvalidAmount
	}

	fl := storage.NewFileLoader(client.GetWalletFilePath())
	wm := client.NewWalletManager(fl)
	err := wm.LoadFromFile()

	if err != nil {
		return &rpcpb.SendResponse{Message: "Error loading local wallets"}, err
	}

	senderWallet := wm.GetWalletByAddress(sendFromAddress)
	if len(senderWallet.Addresses) == 0 {
		return &rpcpb.SendResponse{Message: "Sender wallet not found"}, errors.New("sender address not found in local wallet")
	}

	err = logic.Send(senderWallet, sendToAddress, sendAmount, 0, rpcSerivce.node.GetBlockchain(), rpcSerivce.node)
	if err != nil {
		return &rpcpb.SendResponse{Message: "Error sending"}, err
	}

	return &rpcpb.SendResponse{Message: "Sent"}, nil
}

func (rpcSerivce *RpcService) RpcGetPeerInfo(ctx context.Context, in *rpcpb.GetPeerInfoRequest) (*rpcpb.GetPeerInfoResponse, error) {
	return &rpcpb.GetPeerInfoResponse{
		PeerList: rpcSerivce.node.GetPeerList().ToProto().(*networkpb.Peerlist),
	}, nil
}

func (rpcSerivce *RpcService) RpcGetBlockchainInfo(ctx context.Context, in *rpcpb.GetBlockchainInfoRequest) (*rpcpb.GetBlockchainInfoResponse, error) {
	return &rpcpb.GetBlockchainInfoResponse{
		TailBlockHash: rpcSerivce.node.GetBlockchain().GetTailBlockHash(),
		BlockHeight:   rpcSerivce.node.GetBlockchain().GetMaxHeight(),
	}, nil
}

func (rpcService *RpcService) RpcGetUTXO(ctx context.Context, in *rpcpb.GetUTXORequest) (*rpcpb.GetUTXOResponse, error) {
	utxoIndex := core.LoadUTXOIndex(rpcService.node.GetBlockchain().GetDb())
	publicKeyHash, err = core.Address(in.Address).GetPubKeyHash()
	if err == false {
		return &rpcpb.GetUTXOResponse{ErrorCode: InvalidAddress}, nil
	}

	utxos := utxoIndex.GetUTXOsByPubKey(publicKeyHash)
	response := rpcpb.GetUTXOResponse{ErrorCode: OK}
	for _, utxo := range utxos {
		response.Utxos = append(response.Utxos, &rpcpb.UTXO{utxo.Value.BigInt().Int64(), utxo.PubKeyHash, utxo.Txid, uint32(utxo.TxIndex)})
	}

	return &response, nil
}

func (rpcService *RpcService) RpcGetBlocks(ctx context.Context, in *rpcpb.GetBlocksRequest) (*rpcpb.GetBlocksResponse, error) {
	return &rpcpb.GetBlocksResponse{Message: "Test"}, nil
}
