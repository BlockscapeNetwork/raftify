package raftify

import (
	"encoding/json"
)

// toCandidate initiates the transition into a candidate node for the next term. Calling toCandidate
// on a node that already is in the candidate state just resets the data.
func (n *Node) toCandidate() {
	if n.state == Leader || n.state == Follower {
		n.logger.Println("[WARN] raftify: leader and follower nodes cannot directly switch to candidate")
		return
	}

	n.logger.Printf("[INFO] raftify: Entering candidate state for term %v\n", n.currentTerm+1)
	n.resetTimeout()

	n.currentTerm++
	n.votedFor = n.config.ID
	n.voteList.reset(n.memberlist.Members())
	n.voteList.remove(n.config.ID) // Self vote
	n.state = Candidate

	n.startMessageTicker() // Used to periodically send out vote requests
	n.sendVoteRequestToAll(n.voteList.pending)
}

// runCandidate runs the candidate loop. This function is called within the runLoop function.
func (n *Node) runCandidate() {
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

		case VoteRequestMsg:
			var content VoteRequest
			if err := json.Unmarshal(msg.Content, &content); err != nil {
				n.logger.Printf("[ERR] raftify: error while unmarshaling vote request message: %v\n", err.Error())
				break
			}
			n.handleVoteRequest(content)

		case VoteResponseMsg:
			var content VoteResponse
			if err := json.Unmarshal(msg.Content, &content); err != nil {
				n.logger.Printf("[ERR] raftify: error while unmarshaling vote response message: %v\n", err.Error())
				break
			}
			n.handleVoteResponse(content)

		case NewQuorumMsg:
			var content NewQuorum
			if err := json.Unmarshal(msg.Content, &content); err != nil {
				n.logger.Printf("[ERR] raftify: error while unmarshaling new quorum message: %v\n", err.Error())
				break
			}
			n.handleNewQuorum(content)

		default:
			n.logger.Printf("[WARN] raftify: received %v as candidate, discarding...\n", msg.Type.toString())
		}

	case <-n.messageTicker.C:
		n.sendVoteRequestToAll(n.voteList.pending)

	case <-n.timeoutTimer.C:
		n.logger.Println("[DEBUG] raftify: Election timeout elapsed")

		if n.quorumReached(n.voteList.received) {
			n.logger.Printf("[INFO] raftify: Candidate reached quorum by itself (single-node cluster)")
			n.toLeader()
			return
		}

		n.toCandidate()

	case <-n.events.eventCh:
		n.saveState()

	case <-n.shutdownCh:
		n.toPreShutdown()
	}
}
