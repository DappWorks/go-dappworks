package sendTxFromMiner

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/config"
	configpb "github.com/dappley/go-dappley/config/pb"
	"github.com/dappley/go-dappley/core/account"
	"github.com/dappley/go-dappley/core/transaction"
	transactionpb "github.com/dappley/go-dappley/core/transaction/pb"
	"github.com/dappley/go-dappley/core/utxo"
	"github.com/dappley/go-dappley/logic"
	"github.com/dappley/go-dappley/logic/ltransaction"
	rpcpb "github.com/dappley/go-dappley/rpc/pb"
	"github.com/dappley/go-dappley/wallet"
	logger "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func SendTxFromMiner(ctx context.Context, c interface{},privateKey string,toAddress string,amount int)  {
	var filePath string
	flag.StringVar(&filePath, "file", "miner.conf", "Miner config file path")
	minerConfig := &configpb.MinerConfig{}
	config.LoadConfig(filePath, minerConfig)
	fmt.Println("privateKey:",minerConfig.GetPrivateKey())
	minerAcconut := importWallet(minerConfig.GetPrivateKey(),"1")
	sendTxFromMiner(ctx,c,minerAcconut.GetAddress().String(),toAddress,amount)
}

func importWallet(privateKey string,passphrase string)*account.Account {
	minerAccount := account.NewAccountByPrivateKey(privateKey)
	err := logic.ImportAccountWithPassphrase(minerAccount,passphrase)
	if err != nil {
		logger.Error("Error:", err.Error())
		return nil
	}
	return minerAccount
}

func sendTxFromMiner(ctx context.Context, c interface{},fromAddress string,toAddress string,amount int) {
	var data string
	response, err := logic.GetUtxosAccordingToAmountStream(c.(rpcpb.RpcServiceClient), &rpcpb.GetUTXORequest{
		Address: fromAddress,
		Amount: int64(amount),
	})

	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			logger.Error("Error: server is not reachable!")
		default:
			logger.Error("Error:", status.Convert(err).Message())
		}
		return
	}
	utxos := response.GetUtxos()
	var inputUtxos []*utxo.UTXO
	for _, u := range utxos {
		utxo := utxo.UTXO{}
		utxo.FromProto(u)
		inputUtxos = append(inputUtxos, &utxo)
	}
	tip := common.NewAmount(0)
	gasLimit := common.NewAmount(0)
	gasPrice := common.NewAmount(0)
	tx_utxos, err := GetUTXOsfromAmount(inputUtxos, common.NewAmount(uint64(amount)), tip, gasLimit, gasPrice)
	if err != nil {
		logger.Error("Error:", err.Error())
		return
	}
	am, err := logic.GetAccountManager(wallet.GetAccountFilePath())
	if err != nil {
		logger.Error("Error:", err.Error())
		return
	}
	senderAccount := am.GetAccountByAddress(account.NewAddress(fromAddress))
	if senderAccount == nil {
		logger.Error("Error: invalid account address.")
		return
	}
	sendTxParam := transaction.NewSendTxParam(account.NewAddress(fromAddress), senderAccount.GetKeyPair(),
		account.NewAddress(toAddress), common.NewAmount(uint64(amount)), tip, gasLimit, gasPrice, data)
	tx, err := ltransaction.NewNormalUTXOTransaction(tx_utxos, sendTxParam)
	sendTransactionRequest := &rpcpb.SendTransactionRequest{Transaction: tx.ToProto().(*transactionpb.Transaction)}
	_, err = c.(rpcpb.RpcServiceClient).RpcSendTransaction(ctx, sendTransactionRequest)
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			logger.Error("Error: server is not reachable!")
		default:
			logger.Error("Error:", status.Convert(err).Message())
		}
		return
	}
	logger.Info("Transaction is sent! Pending approval from network.")
}
var (
	ErrInsufficientFund = errors.New("cli: the balance is insufficient")
	ErrTooManyUtxoFund  = errors.New("cli: utxo is too many should to merge")
)
func GetUTXOsfromAmount(inputUTXOs []*utxo.UTXO, amount *common.Amount, tip *common.Amount, gasLimit *common.Amount, gasPrice *common.Amount) ([]*utxo.UTXO, error) {
	if tip != nil {
		amount = amount.Add(tip)
	}
	if gasLimit != nil {
		limitedFee := gasLimit.Mul(gasPrice)
		amount = amount.Add(limitedFee)
	}
	var retUtxos []*utxo.UTXO
	sum := common.NewAmount(0)
	vinRulesCheck := false
	for i := 0; i < len(inputUTXOs); i++ {
		retUtxos = append(retUtxos, inputUTXOs[i])
		sum = sum.Add(inputUTXOs[i].Value)
		if vinRules(sum, amount, i, len(inputUTXOs)) {
			vinRulesCheck = true
			break
		}
	}
	if vinRulesCheck {
		return retUtxos, nil
	}
	if sum.Cmp(amount) > 0 {
		return nil, ErrTooManyUtxoFund
	}
	return nil, ErrInsufficientFund
}

func vinRules(utxoSum, amount *common.Amount, utxoNum, remainUtxoNum int) bool {
	return utxoSum.Cmp(amount) >= 0 && (utxoNum == 50 || remainUtxoNum < 100)
}