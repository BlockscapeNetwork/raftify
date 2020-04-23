package raftify

import (
	"fmt"
	"log"

	"github.com/hashicorp/memberlist"
)

// HeartbeatIDList is a custom type for a list of heartbeat IDs. This is
// needed in order to be able to differentiate between heartbeat responses
// within their respective ticker cycles and those which came from an
// outdated cycle and therefore do not count anymore.
type HeartbeatIDList struct {
	// Superordinate logger of Node struct.
	logger *log.Logger

	// The ID the next heartbeat that is sent out will be identified by.
	// This way, for every term each individual heartbeat can be uniquely
	// identified.
	currentHeartbeatID uint64

	// The number of votes received that granted the node its vote.
	received int

	// The IDs of the heartbeats that have not yet been replied to.
	pending []uint64

	// The number of ticker cycles a leader has not received a majority of
	// heartbeat responses from the other cluster members.
	subQuorumCycles int
}

// add adds a heartbeat ID to the heartbeat ID list. This is used to uniquely
// track all individual heartbeats sent out during a leader quorum cycle.
func (h *HeartbeatIDList) add(heartbeatid uint64) {
	h.pending = append(h.pending, heartbeatid)
}

// remove removes a heartbeat ID from the heartbeat ID list. Returns an error
// if the ID could not be found.
func (h *HeartbeatIDList) remove(heartbeatid uint64) error {
	for i, id := range h.pending {
		if heartbeatid == id {
			h.pending = append(h.pending[:i], h.pending[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("no heartbeats pending from id %v", heartbeatid)
}

// reset resets the heartbeats ID pending and received.
func (h *HeartbeatIDList) reset() {
	h.received = 1
	h.pending = []uint64{} // 1 in order to account for heartbeat from leader to itself
}

// VoteList is a custom type for a list of nodes who haven't (pre)voted yet.
type VoteList struct {
	// Superordinate logger of Node struct.
	logger *log.Logger

	// The number of (pre)votes that have been granted.
	received int

	// The nodes who have not yet replied to a (pre)vote request.
	pending []*memberlist.Node

	// The number of cycles a precandidate has not received any reply
	// to a prevote request. This is used to make a follower turning
	// into a precandidate aware of a network partition. Used only
	// in the precandidate state.
	missedPrevoteCycles int
}

// remove removes the (pre)voter from the list of (pre)votes pending.
func (v *VoteList) remove(voter string) error {
	for i := range v.pending {
		if v.pending[i].Name == voter {
			v.pending = append(v.pending[:i], v.pending[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("no votes pending from %v", voter)
}

// reset resets the votes received and votes pending of the vote list.
// The memberlist which needs to be passed in is the initial list for
// the (pre)votes pending.
func (v *VoteList) reset(initPending []*memberlist.Node) {
	v.received = 1 // 1 in order to account for self prevote/vote
	v.pending = initPending
}
