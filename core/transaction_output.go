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
	"fmt"
	"github.com/dappley/go-dappley/util"
	"github.com/golang/protobuf/proto"
	logger "github.com/sirupsen/logrus"

	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core/pb"
)

type TXOutput struct {
	Value      *common.Amount
	PubKeyHash PubKeyHash
	Contract   string
}

func (out *TXOutput) Lock(address Address) {
	hash, _ := address.GetPubKeyHash()
	out.PubKeyHash = PubKeyHash(hash)
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare([]byte(out.PubKeyHash), pubKeyHash) == 0
}

func NewTXOutput(value *common.Amount, address Address) *TXOutput {
	return NewTxOut(value, address, "")
}

func NewContractTXOutput(address Address, contract string) *TXOutput {
	return NewTxOut(common.NewAmount(0), address, contract)
}

func NewTxOut(value *common.Amount, address Address, contract string) *TXOutput {
	var pubKeyHash PubKeyHash
	txo := &TXOutput{value, pubKeyHash, contract}
	txo.Lock(address)
	return txo
}

func (out *TXOutput) IsFoundInRewardStorage(rewardStorage map[string]string) bool {

	val, isFound := rewardStorage[out.PubKeyHash.GenerateAddress().String()]
	if !isFound {
		return false
	}

	amount, err := common.NewAmountFromString(val)
	if err != nil {
		logger.WithError(err).WithFields(logger.Fields{
			"reward": val,
		}).Warn("TXOutput: Reward amount is in invalid format.")
		return false
	}

	return out.Value.Cmp(amount) == 0
}

func (out *TXOutput) ToProto() proto.Message {
	return &corepb.TXOutput{
		Value:         out.Value.Bytes(),
		PublicKeyHash: []byte(out.PubKeyHash),
		Contract:      out.Contract,
	}
}

func (out *TXOutput) FromProto(pb proto.Message) {
	out.Value = common.NewAmountFromBytes(pb.(*corepb.TXOutput).GetValue())
	out.PubKeyHash = PubKeyHash(pb.(*corepb.TXOutput).GetPublicKeyHash())
	out.Contract = pb.(*corepb.TXOutput).GetContract()
}

func (out *TXOutput) CheckContractSyntax(sc ScEngine) error {
	if out.Contract != "" {

		function, args := util.DecodeScInput(out.Contract)

		totalArgs := util.PrepareArgs(args)
		functionCallScript := prepareFuncCallScript(function, totalArgs)

		if function != "" {
			return sc.CheckContactSyntax(functionCallScript)
		}
		return sc.CheckContactSyntax(out.Contract)
	}
	return nil
}

func prepareFuncCallScript(function, args string) string {
	return fmt.Sprintf(
		`var instance = new _native_require();instance["%s"].apply(instance, [%s]);`,
		function,
		args,
	)
}
