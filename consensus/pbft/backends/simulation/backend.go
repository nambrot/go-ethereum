// Copyright 2017 AMIS Technologies
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package simulation

import (
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/pbft"
	"github.com/ethereum/go-ethereum/consensus/pbft/backends/simple"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
)

const (
	PBFTMsg = 0x11
)

func New(nodeKey *NodeKey, genesis *types.Block) *ProtocolManager {
	eventMux := new(event.TypeMux)
	p := newPeer(nodeKey)
	memDB, _ := ethdb.NewMemDatabase()
	pm := &ProtocolManager{
		rwMutex:   new(sync.RWMutex),
		backend:   simple.New(3000, eventMux, nodeKey.PrivateKey(), memDB),
		genesis:   genesis,
		mockChain: newMockChain(genesis),
		eventMux:  eventMux,
		me:        p,
		peers:     make(map[string]*peer),
		logger:    log.New("backend", "simulated", "id", p.ID()),
	}

	return pm
}

// ----------------------------------------------------------------------------

type ProtocolManager struct {
	rwMutex *sync.RWMutex

	backend   consensus.PBFT
	genesis   *types.Block
	mockChain *mockChain
	eventMux  *event.TypeMux
	eventSub  *event.TypeMuxSubscription
	me        *peer
	peers     map[string]*peer
	logger    log.Logger
}

func (pm *ProtocolManager) Start() {
	pm.eventSub = pm.eventMux.Subscribe(pbft.ConsensusDataEvent{}, pbft.ConsensusCommitBlockEvent{})
	go pm.consensusEventLoop()

	pm.backend.Start(pm.mockChain)
	go pm.handlePeerMessage()
}

func (pm *ProtocolManager) Stop() {
	pm.me.Close()
	pm.backend.Stop()
	pm.eventSub.Unsubscribe()
}

func (pm *ProtocolManager) TryNewRequest() {
	// try to make next block
	pm.newRequest(pm.genesis)
}

func (pm *ProtocolManager) SelfPeer() *peer {
	return pm.me
}

func (pm *ProtocolManager) AddPeer(p *peer) {
	pm.rwMutex.Lock()
	defer pm.rwMutex.Unlock()
	if p.ID() != pm.me.ID() {
		pm.peers[p.ID()] = p
		pm.backend.AddPeer(p.ID(), p.PublicKey())
	}
}

func (pm *ProtocolManager) handlePeerMessage() {
	for {
		payload, err := pm.readPeerMessage()
		if err != nil {
			return
		}

		// FIXME: pass first peer id for test, the real source is hidden in payload
		pm.rwMutex.RLock()
		var randP *peer
		for _, p := range pm.peers {
			randP = p
			break
		}
		pm.rwMutex.RUnlock()
		pm.backend.HandleMsg(randP.ID(), payload)
	}
}

func (pm *ProtocolManager) readPeerMessage() ([]byte, error) {
	m, err := pm.me.ReadMsg()
	if err != nil {
		pm.logger.Error("Failed to ReadMsg", "error", err)
		return nil, err
	}
	defer m.Discard()

	var payload []byte
	err = m.Decode(&payload)
	if err != nil {
		pm.logger.Error("Failed to read payload", "error", err, "msg", m)
		return nil, err
	}
	return payload, nil
}

func (pm *ProtocolManager) consensusEventLoop() {
	for obj := range pm.eventSub.Chan() {
		switch ev := obj.Data.(type) {
		case pbft.ConsensusDataEvent:
			pm.sendEvent(ev)
		case pbft.ConsensusCommitBlockEvent:
			pm.commitBlock(ev.Block)
		}
	}
}

func (pm *ProtocolManager) sendEvent(event pbft.ConsensusDataEvent) {
	pm.rwMutex.RLock()
	defer pm.rwMutex.RUnlock()

	p := pm.peers[event.PeerID]
	if p == nil {
		return
	}
	p2p.Send(p, PBFTMsg, event.Data)
}

func (pm *ProtocolManager) commitBlock(block *types.Block) {
	pm.mockChain.InsertBlock(block)
	pm.newRequest(block)
}

func (pm *ProtocolManager) newRequest(lastBlock *types.Block) {
	block := makeBlock(lastBlock, lastBlock.Number())
	// will get consensus block after Seal if backend is the proposer; otherwise, will get error
	newBlock, err := pm.backend.Seal(pm.mockChain, block, nil)
	if newBlock != nil && err == nil {
		pm.commitBlock(block)
	}
}

func makeBlock(parent *types.Block, num *big.Int) *types.Block {
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     num.Add(num, common.Big1),
		GasLimit:   new(big.Int),
		GasUsed:    new(big.Int),
		Extra:      parent.Extra(),
		Time:       big.NewInt(int64(time.Now().Nanosecond())),
	}

	return types.NewBlockWithHeader(header)
}
