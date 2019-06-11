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

package main

import (
	"flag"

	logger "github.com/sirupsen/logrus"

	"github.com/dappley/go-dappley/config"
	configpb "github.com/dappley/go-dappley/config/pb"
	"github.com/dappley/go-dappley/consensus"
	vm "github.com/dappley/go-dappley/contract"
	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/logic"
	metrics "github.com/dappley/go-dappley/metrics/api"
	"github.com/dappley/go-dappley/network"
	"github.com/dappley/go-dappley/rpc"
	"github.com/dappley/go-dappley/storage"
)

const (
	genesisAddr     = "121yKAXeG4cw6uaGCBYjWk9yTWmMkhcoDD"
	configFilePath  = "conf/default.conf"
	genesisFilePath = "conf/genesis.conf"
	defaultPassword = "password"
	size1kB         = 1024
)

func main() {

	logger.SetFormatter(&logger.TextFormatter{
		FullTimestamp: true,
	})

	logger.SetLevel(logger.InfoLevel)

	var filePath string
	flag.StringVar(&filePath, "f", configFilePath, "Configuration File Path. Default to conf/default.conf")
	flag.Parse()

	//load genesis file information
	genesisConf := &configpb.DynastyConfig{}
	config.LoadConfig(genesisFilePath, genesisConf)

	if genesisConf == nil {
		logger.Error("Cannot load genesis configurations from file! Exiting...")
		return
	}

	//load config file information
	conf := &configpb.Config{}
	config.LoadConfig(filePath, conf)
	if conf == nil {
		logger.Error("Cannot load configurations from file! Exiting...")
		return
	}

	//setup
	db := storage.OpenDatabase(conf.GetNodeConfig().GetDbPath())
	defer db.Close()

	//create blockchain
	conss, dynasty := initConsensus(genesisConf)
	txPoolLimit := conf.GetNodeConfig().GetTxPoolLimit() * size1kB
	nodeAddr := conf.GetNodeConfig().GetNodeAddress()
	blkSizeLimit := conf.GetNodeConfig().GetBlkSizeLimit() * size1kB
	scManager := vm.NewV8EngineManager(core.NewAddress(nodeAddr))
	bc, err := core.GetBlockchain(db, conss, txPoolLimit, scManager, int(blkSizeLimit))
	if err != nil {
		bc, err = logic.CreateBlockchain(core.NewAddress(genesisAddr), db, conss, txPoolLimit, scManager, int(blkSizeLimit))
		if err != nil {
			logger.Panic(err)
		}
	}
	bc.SetState(core.BlockchainInit)

	node, err := initNode(conf, bc, dynasty.IsProducer(conf.GetNodeConfig().GetNodeAddress()))
	if err != nil {
		logger.WithError(err).Error("Failed to initialize the node! Exiting...")
		return
	}
	defer node.Stop()

	bc.SetState(core.BlockchainReady)
	node.DownloadBlocks(bc)

	//start metrics api server
	nodeConf := conf.GetNodeConfig()
	metrics.StartAPI(node, nodeConf.GetMetricsHost(), nodeConf.GetMetricsPort(),
		nodeConf.GetMetricsInterval(), nodeConf.GetMetricsPollingInterval())

	//start rpc server
	server := rpc.NewGrpcServer(node, defaultPassword)
	server.Start(conf.GetNodeConfig().GetRpcPort())
	defer server.Stop()

	//start mining
	minerAddr := conf.GetConsensusConfig().GetMinerAddress()
	conss.Setup(node, minerAddr)
	conss.SetKey(conf.GetConsensusConfig().GetPrivateKey())
	logger.WithFields(logger.Fields{
		"miner_address": minerAddr,
	}).Info("Consensus is configured.")

	logic.SetLockWallet() //lock the wallet
	logic.SetMinerKeyPair(conf.GetConsensusConfig().GetPrivateKey())
	conss.Start()
	defer conss.Stop()

	select {}
}

func initConsensus(conf *configpb.DynastyConfig) (core.Consensus, *consensus.Dynasty) {
	//set up consensus
	conss := consensus.NewDPOS()
	dynasty := consensus.NewDynastyWithConfigProducers(conf.GetProducers(), (int)(conf.GetMaxProducers()))
	conss.SetDynasty(dynasty)
	return conss, dynasty
}

func initNode(conf *configpb.Config, bc *core.Blockchain, isProducer bool) (*network.Node, error) {
	//create node
	node := network.NewNode(bc, core.NewBlockPool(0))
	node.SetIsProducer(isProducer)
	nodeConfig := conf.GetNodeConfig()
	port := nodeConfig.GetPort()
	keyPath := nodeConfig.GetKeyPath()
	if keyPath != "" {
		err := node.LoadNetworkKeyFromFile(keyPath)
		if err != nil {
			logger.Error(err)
		}
	}

	seeds := nodeConfig.GetSeed()
	for _, seed := range seeds {
		node.GetPeerManager().AddSeedByString(seed)
	}

	err := node.Start(int(port))
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	return node, nil
}
