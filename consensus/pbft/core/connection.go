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

package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/pbft"
)

func (c *core) handleConnection(addr common.Address) error {
	logger := c.logger.New("fromAddress", addr.Hex())
	logger.Debug("handleConnection")

	// send hello message
	c.sendHello(addr)
	return nil
}

func (c *core) sendHello(addr common.Address) {
	logger := c.logger.New("toAddress", addr.Hex())
	logger.Debug("sendHello")

	// deliver lastest snapshot (no need to be stable)
	c.snapshotsMu.RLock()
	if len(c.snapshots) == 0 {
		return
	}
	snap := c.snapshots[len(c.snapshots)-1]
	c.snapshotsMu.RUnlock()

	state := pbft.Hello{
		// FIXME: it's better to deliver whole snapshot
		Preprepare: snap.Preprepare,
	}

	c.send(pbft.MsgHello, state, addr)
}

func (c *core) handleHello(hello *pbft.Hello, src pbft.Validator) error {
	if c.state != StateSync {
		return nil
	}
	if err := c.syncs.AddCheckpoint(hello.Preprepare, src.Address()); err != nil {
		return err
	}
	return nil
}
