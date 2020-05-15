package raftify

import (
	"encoding/json"
)

// toFollower initiates the transition into a follower node for a given term. Calling toFollower
// on a node that already is in the follower state just resets the data.
func (n *Node) toFollower(term uint64) {
	n.logger.Printf("[INFO] raftify: Entering follower state for term %v\n", term)

	n.resetTimeout()
	n.messageTicker.Stop() // Stop the ticker if the node was a leader or candidate prior to becoming a follower.

	n.currentTerm = term
	n.votedFor = ""
	n.state = Follower
}

// runFollower runs the follower loop. This function is called within the runLoop function.
func (n *Node) runFollower() {
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

		default:
			n.logger.Printf("[WARN] raftify: received %v as follower, discarding...\n", msg.Type.toString())
		}

	case <-n.timeoutTimer.C:
		n.logger.Println("[DEBUG] raftify: Heartbeat timeout elapsed")
		if n.rejoin {
			// Keep trying to rejoin the cluster.
			if err := n.tryJoin(); err != nil {
				n.logger.Printf("[ERR] raftify: couldn't rejoin cluster: %v\n", err.Error())
				n.resetTimeout()
				break
			}

			// If rejoin was successful, set flag accordingly.
			n.rejoin = false
			n.logger.Println("[DEBUG] raftify: Successfully rejoined the cluster ✓")

			// Signal successful bootstrap and allow InitNode to return.
			n.bootstrapCh <- nil
		}
		n.toPreCandidate()

	case <-n.events.eventCh:
		n.saveState()

	case <-n.shutdownCh:
		n.toShutdown()
	}
}
