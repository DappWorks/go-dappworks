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

package network

import (
	"github.com/dappley/go-dappley/network/network_model"
	"math/rand"
	"sync"
	"time"

	"github.com/dappley/go-dappley/network/pb"
	"github.com/dappley/go-dappley/storage"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/peer"
	logger "github.com/sirupsen/logrus"
)

type ConnectionType int

const (
	syncPeersKey = "SyncPeers"

	maxSyncPeersCount = 32

	syncPeersScheduleTime  = 10 * time.Minute
	PeerConnectionInterval = 15 * time.Minute

	GetPeerListRequest  = "GetPeerListRequest"
	GetPeerListResponse = "GetPeerListResponse"
)

var (
	subscribedTopics = []string{
		GetPeerListRequest,
		GetPeerListResponse,
	}
)

type onPeerListReceived func(newPeers []network_model.PeerInfo)

type PeerManager struct {
	hostPeerId           peer.ID
	seeds                map[peer.ID]network_model.PeerInfo
	syncPeers            map[peer.ID]network_model.PeerInfo
	netService           NetService
	db                   Storage
	onPeerListReceivedCb onPeerListReceived
	mutex                sync.RWMutex
}

//NewPeerManager create a new peer manager object
func NewPeerManager(netService NetService, db Storage, onPeerListReceivedCb onPeerListReceived, seeds []string) *PeerManager {
	pm := &PeerManager{
		seeds:                make(map[peer.ID]network_model.PeerInfo),
		syncPeers:            make(map[peer.ID]network_model.PeerInfo),
		mutex:                sync.RWMutex{},
		netService:           netService,
		db:                   db,
		onPeerListReceivedCb: onPeerListReceivedCb,
	}
	pm.Subscribe()
	pm.addSeeds(seeds)
	if db != nil {
		pm.loadSyncPeers()
	}
	return pm
}

//GetSeeds return a slice of seed peers
func (pm *PeerManager) GetSeeds() []network_model.PeerInfo {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	allSeeds := []network_model.PeerInfo{}
	for _, seed := range pm.seeds {
		allSeeds = append(allSeeds, seed)
	}
	return allSeeds
}

//GetSyncPeers return a slice of sync peers
func (pm *PeerManager) GetSyncPeers() []network_model.PeerInfo {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	allPeers := []network_model.PeerInfo{}
	for _, peer := range pm.syncPeers {
		allPeers = append(allPeers, peer)
	}
	return allPeers
}

//GetSubscribedTopics returns subscribed topics
func (pm *PeerManager) Subscribe() {
	if pm.netService == nil {
		return
	}

	for _, topic := range subscribedTopics {
		pm.netService.Subscribe(topic, pm.GetCommandHandler(topic))
	}
}

//GetCommandHandler returns the corresponding command handler
func (pm *PeerManager) GetCommandHandler(cmd string) network_model.CommandHandlerFunc {
	switch cmd {
	case GetPeerListRequest:
		return pm.GetPeerListRequestHandler
	case GetPeerListResponse:
		return pm.GetPeerListResponseHandler
	}
	return nil
}

//addSeeds add seed peers
func (pm *PeerManager) addSeeds(seeds []string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	for _, seed := range seeds {
		pm.addSeedByString(seed)
	}
}

//addSeedByString adds seed peer by multiaddr string
func (pm *PeerManager) addSeedByString(fullAddr string) {

	peerInfo, err := network_model.NewPeerInfoFromString(fullAddr)
	if err != nil {
		logger.WithError(err).WithFields(logger.Fields{
			"full_addr": fullAddr,
		}).Warn("PeerManager: create PeerInfo failed.")
	}

	pm.seeds[peerInfo.PeerId] = peerInfo
}

//AddSeedByPeerInfo adds seed by peerInfo
func (pm *PeerManager) AddSeedByPeerInfo(peerInfo network_model.PeerInfo) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if peerInfo.PeerId.String() == "" {
		return
	}

	pm.seeds[peerInfo.PeerId] = peerInfo

}

//UpdateSyncPeers synchronizes the sync peers with the connected peer list
func (pm *PeerManager) UpdateSyncPeers(connectedPeerList map[peer.ID]network_model.PeerInfo) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.syncPeers = connectedPeerList
	for _, peer := range pm.syncPeers {
		if _, isSeed := pm.seeds[peer.PeerId]; isSeed {
			delete(pm.syncPeers, peer.PeerId)
		}
	}

	pm.saveSyncPeers()
}

//AddSyncPeer adds a sync peer and saves it to database
func (pm *PeerManager) AddSyncPeer(peer network_model.PeerInfo) {

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.addSyncPeer(peer)
	pm.saveSyncPeers()
}

//addSyncPeer adds a sync peer
func (pm *PeerManager) addSyncPeer(peer network_model.PeerInfo) {

	if pm.isPeerNew(peer.PeerId) {
		pm.syncPeers[peer.PeerId] = peer
	}

}

//Start starts peer manager thread
func (pm *PeerManager) Start() {
	pm.startSyncPeersSchedule()
}

func (pm *PeerManager) SetHostPeerId(hostPeerId peer.ID) {
	pm.hostPeerId = hostPeerId
}

//isPeerExisted returns if a peer exists
func (pm *PeerManager) isPeerExisted(peerId peer.ID) bool {

	if _, existed := pm.seeds[peerId]; existed {
		return true
	}

	if _, existed := pm.syncPeers[peerId]; existed {
		return true
	}

	return false
}

func (pm *PeerManager) isPeerNew(peerId peer.ID) bool {
	return !pm.isPeerExisted(peerId) && peerId != pm.hostPeerId
}

//AddPeers adds one of its peers' peerlist to its own peerlist
func (pm *PeerManager) AddPeers(peers []network_model.PeerInfo) {

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	for _, peer := range peers {
		pm.addSyncPeer(peer)
	}
	pm.saveSyncPeers()

}

func (pm *PeerManager) GetAllPeers() []network_model.PeerInfo {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var allPeers []network_model.PeerInfo
	for _, seed := range pm.seeds {
		allPeers = append(allPeers, seed)
	}
	for _, peer := range pm.syncPeers {
		allPeers = append(allPeers, peer)
	}
	return allPeers
}

//GetRandomPeers get a number of random connected peers
func (pm *PeerManager) GetRandomPeers(numOfPeers int) []network_model.PeerInfo {

	peers := pm.GetAllPeers()

	logger.WithFields(logger.Fields{
		"numOfPeersRequested": numOfPeers,
		"numOfAllPeers":       len(peers),
	}).Info("PeerManager: GetRandomPeers")

	if numOfPeers >= len(peers) {
		return peers
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})

	return peers[0:numOfPeers]
}

//startSyncPeersSchedule starts synchronize peer list with its peers
func (pm *PeerManager) startSyncPeersSchedule() {

	go func() {
		pm.BroadcastGetPeerListRequest()
		ticker := time.NewTicker(syncPeersScheduleTime)
		for {
			select {
			case <-ticker.C:
				pm.BroadcastGetPeerListRequest()
			}
		}
	}()
}

//saveSyncPeers saves the syncPeers to database
func (pm *PeerManager) saveSyncPeers() {

	var peerPbs []*networkpb.PeerInfo
	for _, peerInfo := range pm.syncPeers {
		peerPbs = append(peerPbs, peerInfo.ToProto().(*networkpb.PeerInfo))
	}

	bytes, err := proto.Marshal(&networkpb.ReturnPeerList{PeerList: peerPbs})
	if err != nil {
		logger.WithError(err).Info("PeerManager: serialize sync peers failed.")
	}

	err = pm.db.Put([]byte(syncPeersKey), bytes)
	if err != nil {
		logger.WithError(err).Info("PeerManager: save sync peers failed.")
	}
}

//loadSyncPeers loads the syncPeers from database
func (pm *PeerManager) loadSyncPeers() error {

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	peersBytes, err := pm.db.Get([]byte(syncPeersKey))

	if err != nil {
		logger.WithError(err).Warn("PeerManager: load sync peers database failed.")
		if err == storage.ErrKeyInvalid {
			return nil
		}
		return err
	}

	peerListPb := &networkpb.ReturnPeerList{}

	if err := proto.Unmarshal(peersBytes, peerListPb); err != nil {
		logger.WithError(err).Warn("PeerManager: parse Peerlist failed.")
	}

	for _, peerPb := range peerListPb.GetPeerList() {
		peerInfo := network_model.PeerInfo{}
		if err := peerInfo.FromProto(peerPb); err != nil {
			logger.WithError(err).Warn("PeerManager: parse PeerInfo failed.")
		}

		pm.syncPeers[peerInfo.PeerId] = peerInfo
		logger.WithFields(logger.Fields{
			"peer_id": peerInfo.PeerId,
			"addr":    peerInfo.Addrs[0].String(),
		}).Info("loadSyncPeers")
	}

	logger.WithError(err).Warnf("PeerManager: load sync peers count %v.", len(pm.syncPeers))

	return nil
}

//BroadcastGetPeerListRequest broadcasts a syncPeer request command to all its peers
func (pm *PeerManager) BroadcastGetPeerListRequest() {

	logger.Info("PeerManager: BroadcastGetPeerListRequest")
	getPeerListPb := &networkpb.GetPeerList{
		MaxNumber: int32(maxSyncPeersCount),
	}

	var destination peer.ID
	pm.netService.SendCommand(GetPeerListRequest, getPeerListPb, destination, network_model.Broadcast, network_model.HighPriorityCommand)

}

//SendGetPeerListResponse sends its peer list to destination peer
func (pm *PeerManager) SendGetPeerListResponse(maxNumOfPeers int, destination peer.ID) {

	peers := pm.GetRandomPeers(maxNumOfPeers)
	var peerPbs []*networkpb.PeerInfo
	for i, peerInfo := range peers {
		logger.WithFields(logger.Fields{
			"index":  i,
			"peerId": peerInfo.PeerId,
		}).Info("PeerManager: SendGetPeerListResponse Peer information")
		peerPbs = append(peerPbs, peerInfo.ToProto().(*networkpb.PeerInfo))
	}

	peerList := &networkpb.ReturnPeerList{PeerList: peerPbs}

	pm.netService.SendCommand(GetPeerListResponse, peerList, destination, network_model.Unicast, network_model.HighPriorityCommand)

}

//GetPeerListRequestHandler is the handler to GetPeerListRequest
func (pm *PeerManager) GetPeerListRequestHandler(command *network_model.DappRcvdCmdContext) {
	logger.Info("PeerManager: GetPeerListRequestHandler")

	getPeerlistRequest := &networkpb.GetPeerList{}

	//unmarshal byte to proto
	if err := proto.Unmarshal(command.GetData(), getPeerlistRequest); err != nil {
		logger.WithError(err).Warn("Node: parse GetPeerListRequest failed.")
	}

	pm.SendGetPeerListResponse(int(getPeerlistRequest.GetMaxNumber()), command.GetSource())
}

//GetPeerListResponseHandler is the handler to SendGetPeerListResponse
func (pm *PeerManager) GetPeerListResponseHandler(command *network_model.DappRcvdCmdContext) {

	peerlistPb := &networkpb.ReturnPeerList{}

	if err := proto.Unmarshal(command.GetData(), peerlistPb); err != nil {
		logger.WithError(err).Warn("PeerManager: parse Peerlist failed.")
	}

	var peers []network_model.PeerInfo
	for _, peerPb := range peerlistPb.GetPeerList() {
		peerInfo := network_model.PeerInfo{}
		if err := peerInfo.FromProto(peerPb); err != nil {
			logger.WithError(err).Warn("PeerManager: parse PeerInfo failed.")
		}
		peers = append(peers, peerInfo)
	}

	newPeers := []network_model.PeerInfo{}
	for _, peer := range peers {
		if pm.isPeerNew(peer.PeerId) {
			newPeers = append(newPeers, peer)
		}
	}

	pm.AddPeers(newPeers)
	go pm.onPeerListReceivedCb(newPeers)

}
