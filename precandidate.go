package raftify

import (
	"encoding/json"
)

// toPreCandidate initiates the transition of a follower into a precandidate.
func (n *Node) toPreCandidate() {
	n.logger.Printf("[DEBUG] raftify: Entering precandidate state for term %v", n.currentTerm+1)

	n.resetTimeout()
	n.preVoteList.reset(n.memberlist.Members())
	n.preVoteList.remove(n.config.ID) // Self prevote
	n.state = PreCandidate

	n.sendPreVoteRequestToAll()
}

// runPreCandidate runs the precandidate loop. This function is called within the runLoop function.
func (n *Node) runPreCandidate() {
	select {
	case msgBytes := <-n.messages.messageCh:
		var msg Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			n.logger.Printf("[ERR] raftify: error while unmarshaling wrapper message: %v\n", err.Error())
			break
		}

		switch msg.Type {
		case HeartbeatMsg:
			var content Heartbeat
			if err := json.Unmarshal(msg.Content, &content); err != nil {
				n.logger.Printf("[ERR] raftify: error while unmarshaling heartbeat message: %v\n", err.Error())
				break
			}
			n.handleHeartbeat(content)

		case PreVoteRequestMsg:
			var content PreVoteRequest
			if err := json.Unmarshal(msg.Content, &content); err != nil {
				n.logger.Printf("[ERR] raftify: error while unmarshaling prevote request message: %v\n", err.Error())
				break
			}
			n.handlePreVoteRequest(content)

		case PreVoteResponseMsg:
			var content PreVoteResponse
			if err := json.Unmarshal(msg.Content, &content); err != nil {
				n.logger.Printf("[ERR] raftify: error while unmarshaling prevote response message: %v\n", err.Error())
				break
			}
			n.handlePreVoteResponse(content)

		case VoteRequestMsg:
			var content VoteRequest
			if err := json.Unmarshal(msg.Content, &content); err != nil {
				n.logger.Printf("[ERR] raftify: error while unmarshaling vote request message: %v\n", err.Error())
				break
			}
			n.handleVoteRequest(content)

		default:
			n.logger.Printf("[WARN] raftify: received %v as precandidate, discarding...\n", msg.Type.toString())
		}

	case <-n.timeoutTimer.C:
		n.logger.Println("[DEBUG] raftify: Election timeout elapsed")

		// This is mainly to initiate a quorum check for single-node clusters since checks
		// are done on receival of a vote by default. This happens, for example, if expect is
		// set to 1.
		if n.quorumReached(n.preVoteList.received) {
			n.logger.Printf("[INFO] raftify: PreCandidate reached quorum by itself (single-node cluster)")
			n.toCandidate()
			return
		}

		n.toPreCandidate()

		// If only a minorty keeps prevoting and the precandidate quorum cannot be
		// reached, increment the counter of missed prevote cycles.
		n.preVoteList.missedPrevoteCycles++

		// If the precandidate hasn't reached the prevote quorum too many cycles in a row
		// it assumes it is partitioned out of the main cluster. In that case, it stops
		// collecting prevotes and becomes a follower again where the rejoin flag will
		// trigger a rejoin event. The node will not be able to continue operation until
		// it successfully rejoined the cluster.
		if n.preVoteList.missedPrevoteCycles >= 5 {
			n.logger.Printf("[DEBUG] raftify: %v prevote cycles have passed without any response, preparing rejoin...\n", n.preVoteList.missedPrevoteCycles)
			n.preVoteList.missedPrevoteCycles = 0
			n.toRejoin()
		}

	case <-n.events.eventCh:
		n.saveState()

	case <-n.shutdownCh:
		n.toShutdown()
	}
}
