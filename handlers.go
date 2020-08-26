package raftify

import (
	"fmt"

	"github.com/hashicorp/memberlist"
)

// handleHeartbeat handles the receival of a heartbeat message from a leader.
func (n *Node) handleHeartbeat(msg Heartbeat) {
	// Adjust quorum in any case
	n.quorum = msg.Quorum

	switch n.state {
	case Follower:
		if n.currentTerm < msg.Term {
			n.logger.Printf("[DEBUG] raftify: Received heartbeat with higher term from %v, adopting term %v...\n", msg.LeaderID, msg.Term)
			n.toFollower(msg.Term)
			break
		} else if n.currentTerm > msg.Term {
			n.logger.Printf("[DEBUG] raftify: Received outdated heartbeat from %v, skipping...\n", msg.LeaderID)
			n.sendHeartbeatResponse(msg.LeaderID, msg.HeartbeatID)
			break
		}

		n.logger.Printf("[DEBUG] raftify: Received heartbeat from %v\n", msg.LeaderID)
		n.sendHeartbeatResponse(msg.LeaderID, msg.HeartbeatID)
		n.resetTimeout()

	case PreCandidate:
		if n.currentTerm <= msg.Term {
			n.logger.Printf("[DEBUG] raftify: Received heartbeat with same/higher term from %v, adopting term %v...\n", msg.LeaderID, msg.Term)
			n.toFollower(msg.Term)
			n.sendHeartbeatResponse(msg.LeaderID, msg.HeartbeatID)
			break
		}

		n.logger.Printf("[DEBUG] raftify: Received outdated heartbeat from %v, skipping...\n", msg.LeaderID)

	case Candidate:
		if n.currentTerm <= msg.Term {
			n.logger.Printf("[DEBUG] raftify: Received heartbeat with same/higher term from %v, adopting term %v...\n", msg.LeaderID, msg.Term)
			n.toFollower(msg.Term)
		} else {
			n.logger.Printf("[DEBUG] raftify: Received outdated heartbeat from %v, skipping...\n", msg.LeaderID)
		}

		n.sendHeartbeatResponse(msg.LeaderID, msg.HeartbeatID)

	case Leader:
		panic(fmt.Sprintf("leader %v (term: %v) received heartbeat from %v (term: %v), possible double-signing\n", n.config.ID, n.currentTerm, msg.LeaderID, msg.Term))
	}
}

// handleHeartbeatResponse handles the receival of a heartbeat response message
// from a follower.
func (n *Node) handleHeartbeatResponse(msg HeartbeatResponse) {
	if n.state != Leader {
		n.logger.Printf("[WARN] raftify: received heartbeat response as %v\n", n.state.toString())
		return
	}

	if n.currentTerm < msg.Term {
		n.logger.Printf("[DEBUG] raftify: Received heartbeat response with higher term from %v, skipping...\n", msg.FollowerID)
		n.toFollower(msg.Term)
		return
	} else if n.currentTerm > msg.Term {
		n.logger.Printf("[DEBUG] raftify: Received outdated heartbeat response from %v, skipping...\n", msg.FollowerID)
		return
	}

	n.logger.Printf("[DEBUG] raftify: Received heartbeat response from %v\n", msg.FollowerID)

	// If there are no heartbeats pending from the follower (and he thus cannot be removed)
	// ignore the heartbeat response.
	if err := n.heartbeatIDList.remove(msg.HeartbeatID); err != nil {
		n.logger.Printf("[DEBUG] raftify: %v\n", err.Error())
		return
	}
	n.heartbeatIDList.received++
}

// handlePreVoteRequest handles the receival of a prevote request message from
// a precandidate.
func (n *Node) handlePreVoteRequest(msg PreVoteRequest) {
	n.logger.Printf("[DEBUG] raftify: Received prevote request from %v\n", msg.PreCandidateID)
	if n.state != PreCandidate {
		n.logger.Printf("[WARN] raftify: received prevote request as %v\n", n.state.toString())
		n.sendPreVoteResponse(msg.PreCandidateID, false)
		return
	}

	if n.currentTerm >= msg.NextTerm {
		n.logger.Printf("[DEBUG] raftify: Received outdated prevote request from %v, skipping...\n", msg.PreCandidateID)
		n.sendPreVoteResponse(msg.PreCandidateID, false)
		return
	}
	n.sendPreVoteResponse(msg.PreCandidateID, true)
}

// handlePreVoteResponse handles the receival of a prevote response message from
// a follower.
func (n *Node) handlePreVoteResponse(msg PreVoteResponse) {
	if n.state != PreCandidate {
		n.logger.Printf("[WARN] raftify: received prevote response as %v\n", n.state.toString())
		return
	}

	if n.currentTerm < msg.Term {
		n.logger.Printf("[DEBUG] raftify: Received prevote response with higher term from %v, adopting term %v...\n", msg.FollowerID, msg.Term)
		n.toFollower(msg.Term)
		return
	} else if n.currentTerm > msg.Term {
		n.logger.Printf("[DEBUG] raftify: Received outdated prevote response from %v, skipping...\n", msg.FollowerID)
		return
	}

	// If a prevote was received, reset the missed cycle counter. This counter is
	// used to make a follower node that turns into a precandidate aware of a network
	// partition.
	n.preVoteList.missedPrevoteCycles = 0

	// If there are no prevotes pending from the follower (and he thus cannot be removed)
	// ignore the prevote response.
	if err := n.preVoteList.remove(msg.FollowerID); err != nil {
		n.logger.Printf("[ERR] raftify: %v has already prevoted for %v since the last timeout\n", msg.FollowerID, n.config.ID)
		return
	}

	if msg.PreVoteGranted {
		n.logger.Printf("[DEBUG] raftify: Received prevote response from %v (granted)\n", msg.FollowerID)
		n.preVoteList.received++

		if n.quorumReached(n.preVoteList.received) {
			n.toCandidate()
		}
	} else {
		n.logger.Printf("[DEBUG] raftify: Received prevote response from %v (not granted)\n", msg.FollowerID)
	}
}

// handleVoteRequest handles the receival of a vote request message from a candidate.
func (n *Node) handleVoteRequest(msg VoteRequest) {
	if n.currentTerm < msg.Term {
		n.logger.Printf("[DEBUG] raftify: Received vote request with higher term from %v, adopting term %v...\n", msg.CandidateID, msg.Term)
		n.toFollower(msg.Term)
	} else if n.currentTerm > msg.Term {
		n.logger.Printf("[DEBUG] raftify: Received outdated vote request from %v, skipping...\n", msg.CandidateID)
		n.sendVoteResponse(msg.CandidateID, false)
		return
	} else {
		n.logger.Printf("[DEBUG] raftify: Received vote request from %v\n", msg.CandidateID)
	}

	if n.votedFor != "" {
		n.sendVoteResponse(msg.CandidateID, false)
		return
	}
	n.sendVoteResponse(msg.CandidateID, true)
}

// handleVoteResponse handles the receival of a vote response message from a follower.
func (n *Node) handleVoteResponse(msg VoteResponse) {
	if n.state != Candidate {
		n.logger.Printf("[WARN] raftify: received vote response as %v\n", n.state.toString())
		return
	}

	if n.currentTerm < msg.Term {
		n.logger.Printf("[WARN] raftify: received vote response with higher term from %v, skipping...\n", msg.FollowerID)
		return
	} else if n.currentTerm > msg.Term {
		n.logger.Printf("[DEBUG] raftify: Received outdated vote response from %v, skipping...\n", msg.FollowerID)
		return
	}

	// If there are no votes pending from the follower (and he thus cannot be removed)
	// ignore the vote response.
	if err := n.voteList.remove(msg.FollowerID); err != nil {
		n.logger.Printf("[ERR] raftify: %v\n", err.Error())
		return
	}

	if msg.VoteGranted {
		n.logger.Printf("[DEBUG] raftify: Received vote response from %v (granted)\n", msg.FollowerID)
		n.voteList.received++

		if n.quorumReached(n.voteList.received) {
			n.toLeader()
		}
	} else {
		n.logger.Printf("[DEBUG] raftify: Received vote response from %v (not granted)\n", msg.FollowerID)
	}
}

// handleNewQuorum handles the receival of a new quorum message from a node in the PreShutdown state.
func (n *Node) handleNewQuorum(msg NewQuorum) {
	n.logger.Printf("[DEBUG] raftify: Received new quorum, waiting for %v to leave...\n", msg.LeavingID)

	// If the event is not the leave event fired by the node that announced its exit, do nothing
	if event := <-n.events.eventCh; event.Node.Name != msg.LeavingID && event.Event != memberlist.NodeLeave {
		return
	}

	n.logger.Printf("[DEBUG] raftify: Setting the quorum from %v to %v\n", n.quorum, msg.NewQuorum)
	n.quorum = msg.NewQuorum
	n.saveState()

	if msg.NewQuorum == 1 {
		n.logger.Printf("[DEBUG] raftify: %v is the only node left in the cluster, entering leader state for term %v...", n.config.ID, n.currentTerm)

		// Switch to the Leader state without calling toLeader in order to bypass the state change
		// restriction in this corner case.
		n.timeoutTimer.Stop()  // Leaders have no timeout
		n.startMessageTicker() // Used to periodically send out heartbeat messages
		n.heartbeatIDList.reset()

		n.votedFor = ""
		n.state = Leader
	}
}
