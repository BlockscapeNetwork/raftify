package raftify

import (
	"encoding/json"
)

// toLeader initiates the transition into a leader node. Calling toLeader on a node that already is
// in the leader state just resets the data.
func (n *Node) toLeader() {
	if n.state == Follower {
		n.logger.Println("[WARN] raftify: follower nodes cannot directly switch to leader")
		return
	}

	n.logger.Printf("[INFO] raftify: Entering leader state for term %v\n", n.currentTerm)
	n.timeoutTimer.Stop()  // Leaders have no timeout
	n.startMessageTicker() // Used to periodically send out heartbeat messages
	n.heartbeatIDList.reset()

	n.votedFor = ""
	n.state = Leader

	n.sendHeartbeatToAll()
}

// runLeader runs the leader loop. This function is called within the runLoop function.
func (n *Node) runLeader() {
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

		case HeartbeatResponseMsg:
			var content HeartbeatResponse
			if err := json.Unmarshal(msg.Content, &content); err != nil {
				n.logger.Printf("[ERR] raftify: error while unmarshaling heartbeat response message: %v\n", err.Error())
				break
			}
			n.handleHeartbeatResponse(content)

		case PreVoteRequestMsg:
			var content PreVoteRequest
			if err := json.Unmarshal(msg.Content, &content); err != nil {
				n.logger.Printf("[ERR] raftify: error while unmarshaling prevote request message: %v\n", err.Error())
				break
			}
			n.handlePreVoteRequest(content)

		case VoteRequestMsg:
			var content VoteRequest
			if err := json.Unmarshal(msg.Content, &content); err != nil {
				n.logger.Printf("[ERR] raftify: error while unmarshaling vote request message: %v\n", err.Error())
				break
			}
			n.handleVoteRequest(content)

		default:
			n.logger.Printf("[WARN] raftify: received %v as leader, discarding...\n", msg.Type.toString())
		}

	case <-n.messageTicker.C:
		if !n.quorumReached(n.heartbeatIDList.received) {
			n.heartbeatIDList.subQuorumCycles++
			n.logger.Printf("[DEBUG] raftify: Not enough heartbeat responses for %v cycles\n", n.heartbeatIDList.subQuorumCycles)

			if n.heartbeatIDList.subQuorumCycles >= MaxSubQuorumCycles {
				n.logger.Println("[DEBUG] raftify: Too many cycles without reaching leader quorum, stepping down as a leader...")

				// If at any point the leader doesn't receive enough heartbeat responses anymore
				// it is safe to assume it has been partitioned out into a smaller sub-cluster.
				// It therefore needs to try to rejoin the cluster in order to receive the latest
				// memberlist in case anything has changed during its absence in the other
				// sub-cluster.
				n.rejoin = true

				// Reset heartbeat and quorum counter.
				n.heartbeatIDList.currentHeartbeatID = 0
				n.heartbeatIDList.subQuorumCycles = 0

				// Reload the config so that the memberlist from the state.json is loaded into
				// the peerlist.
				if err := n.loadConfig(); err != nil {
					n.logger.Printf("[ERR] raftify: %v; fallback with peerlist from raftify.json\n", err.Error())
				}

				// Step down as a leader if too many cycles have passed without reaching quorum.
				n.toFollower(n.currentTerm)
				break
			}
		} else {
			// If the quorum was reached, reset the count for how many sub quorum cycles have passed.
			n.heartbeatIDList.subQuorumCycles = 0
		}

		n.sendHeartbeatToAll()

	case <-n.events.eventCh:
		n.saveState()

	case <-n.shutdownCh:
		n.toShutdown()
	}
}
