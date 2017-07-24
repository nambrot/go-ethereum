// Copyright 2017 The go-ethereum Authors
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

package istanbul

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ConsensusDataEvent is posted when sending Istanbul consensus data
type ConsensusDataEvent struct {
	// targets to send message
	Targets map[common.Address]bool
	// consensus message data
	Data []byte
}

// NewCommittedEvent is posted when I'm not a proposer but
// a block has been committed from Istanbul consensus.
type NewCommittedEvent struct {
	Block *types.Block
}

// RequestEvent is posted to propose a proposal
type RequestEvent struct {
	Proposal Proposal
}

// MessageEvent is posted for Istanbul engine communication
type MessageEvent struct {
	Payload []byte
}

// FinalCommittedEvent is posted when a proposal is committed
type FinalCommittedEvent struct {
	Proposal Proposal
	Proposer common.Address
}
