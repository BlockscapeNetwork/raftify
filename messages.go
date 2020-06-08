package raftify

import (
	"encoding/json"

	"github.com/hashicorp/memberlist"
)

// Message is a wrapper struct for all messages used to determine the message type.
type Message struct {
	Type    MessageType     `json:"type"`
	Content json.RawMessage `json:"content"`
}

// Invalidates checks if enqueuing the current broadcast invalidates a previous broadcast.
func (b *Message) Invalidates(other memberlist.Broadcast) bool {
	return false
}

// Finished is invoked when the message will no longer be broadcasted, either due to invalidation
// or to the transmit limit being reached.
func (b *Message) Finished() {}

// UniqueBroadcast is just a marker method for the UniqueBroadcast interface.
func (b *Message) UniqueBroadcast() {}

// Message returns a byte form of the broadcasted message.
func (b *Message) Message() []byte {
	if data, err := json.Marshal(b); err == nil {
		return data
	}
	return []byte("")
}

// Heartbeat defines the message sent out by the leader to all cluster members.
type Heartbeat struct {
	Term        uint64 `json:"term"`
	HeartbeatID uint64 `json:"heartbeat_id"`
	LeaderID    string `json:"leader_id"`
}

// HeartbeatResponse defines the response of a follower to a leader's heartbeat message.
type HeartbeatResponse struct {
	Term        uint64 `json:"term"`
	HeartbeatID uint64 `json:"heartbeat_id"`
	FollowerID  string `json:"follower_id"`
}

// PreVoteRequest defines the message sent out by a follower who is about to become a
// candidate in order to check whether there truly isn't a leader anymore.
type PreVoteRequest struct {
	NextTerm       uint64 `json:"next_term"`
	PreCandidateID string `json:"pre_candidate_id"`
}

// PreVoteResponse defines the response of a follower to a candidate-to-be's pre vote
// request.
type PreVoteResponse struct {
	Term           uint64 `json:"term"`
	FollowerID     string `json:"follower_id"`
	PreVoteGranted bool   `json:"pre_vote_granted"`
}

// VoteRequest defines the message sent out by a candidate to all cluster members to ask
// for votes in order to become leader.
type VoteRequest struct {
	Term        uint64 `json:"term"`
	CandidateID string `json:"candidate_id"`
}

// VoteResponse defines the response of a follower to a candidate's vote request message.
type VoteResponse struct {
	Term        uint64 `json:"term"`
	FollowerID  string `json:"follower_id"`
	VoteGranted bool   `json:"vote_granted"`
}

// IntentionalLeave defines the broadcast message which is sent out on an intentional
// leave event. This message is not to be broadcasted on unintentional/crash-related events.
// The broadcast message is sent out via gossip, so it can be sent out after the node leaves
// but hasn't shut down its listeners yet.
type IntentionalLeave struct {
	NodeEvent memberlist.NodeEventType `json:"node_event"`
	NewQuorum int                      `json:"new_quorum"`
}

// Invalidates checks if enqueuing the current broadcast invalidates a previous broadcast.
func (b *IntentionalLeave) Invalidates(other memberlist.Broadcast) bool {
	return false
}

// Finished is invoked when the message will no longer be broadcasted, either due to invalidation
// or to the transmit limit being reached.
func (b *IntentionalLeave) Finished() {}

// UniqueBroadcast is just a marker method for the UniqueBroadcast interface.
func (b *IntentionalLeave) UniqueBroadcast() {}

// Message returns a byte form of the broadcasted message.
func (b *IntentionalLeave) Message() []byte {
	if data, err := json.Marshal(b); err == nil {
		return data
	}
	return []byte("")
}

// sendHeartbeatToAll sends a heartbeat message to all the other cluster members.
func (n *Node) sendHeartbeatToAll() {
	n.heartbeatIDList.reset()

	hb := Heartbeat{
		HeartbeatID: n.heartbeatIDList.currentHeartbeatID,
		Term:        n.currentTerm,
		LeaderID:    n.config.ID,
	}

	for _, member := range n.memberlist.Members() {
		if member.Name == n.config.ID {
			continue
		}

		hbBytes, _ := json.Marshal(hb)
		msgBytes, _ := json.Marshal(Message{
			Type:    HeartbeatMsg,
			Content: hbBytes,
		})

		if err := n.memberlist.SendBestEffort(member, msgBytes); err != nil {
			n.logger.Printf("[ERR] raftify: couldn't send heartbeat to %v: %v\n", member.Name, err.Error())
			continue
		}

		n.heartbeatIDList.add(hb.HeartbeatID)
		n.heartbeatIDList.currentHeartbeatID++
		hb.HeartbeatID = n.heartbeatIDList.currentHeartbeatID

		n.logger.Printf("[DEBUG] raftify: Sent heartbeat to %v\n", member.Name)
	}
}

// sendHeartbeatResponse sends a heartbeat response message back to the leader it came from.
func (n *Node) sendHeartbeatResponse(leaderid string, heartbeatid uint64) {
	hbRespBytes, _ := json.Marshal(HeartbeatResponse{
		HeartbeatID: heartbeatid,
		Term:        n.currentTerm,
		FollowerID:  n.config.ID,
	})
	msgBytes, _ := json.Marshal(Message{
		Type:    HeartbeatResponseMsg,
		Content: hbRespBytes,
	})

	leaderNode, err := n.getNodeByName(leaderid)
	if err != nil {
		n.logger.Printf("[ERR] raftify: %v\n", err.Error())
		return
	}

	if err := n.memberlist.SendBestEffort(leaderNode, msgBytes); err != nil {
		n.logger.Printf("couldn't send heartbeat response to %v: %v\n", leaderid, err.Error())
		return
	}
	n.logger.Printf("[DEBUG] raftify: Sent heartbeat response to %v\n", leaderid)
}

// sendPreVoteRequestToAll sends a pre vote request message to all cluster members.
// If the node has a state.json it will use prevote requests as a means to rejoin the cluster.
func (n *Node) sendPreVoteRequestToAll() {
	if n.rejoin {
		n.tryJoin()
	}

	reqBytes, _ := json.Marshal(PreVoteRequest{
		NextTerm:       n.currentTerm + 1,
		PreCandidateID: n.config.ID,
	})
	msgBytes, _ := json.Marshal(Message{
		Type:    PreVoteRequestMsg,
		Content: reqBytes,
	})

	for _, member := range n.preVoteList.pending {
		if err := n.memberlist.SendBestEffort(member, msgBytes); err != nil {
			n.logger.Printf("[ERR] raftify: couldn't send prevote request to %v: %v\n", member.Name, err.Error())
			continue
		}
		n.logger.Printf("[DEBUG] raftify: Sent prevote request to %v\n", member.Name)
	}
}

// sendPreVoteResponse sends a prevote response message to the precandidate.
func (n *Node) sendPreVoteResponse(precandidateid string, grant bool) {
	respBytes, _ := json.Marshal(PreVoteResponse{
		Term:           n.currentTerm,
		FollowerID:     n.config.ID,
		PreVoteGranted: grant,
	})
	msgBytes, _ := json.Marshal(Message{
		Type:    PreVoteResponseMsg,
		Content: respBytes,
	})

	precandidateNode, err := n.getNodeByName(precandidateid)
	if err != nil {
		n.logger.Printf("[ERR] raftify: %v\n", err.Error())
		return
	}

	if err := n.memberlist.SendBestEffort(precandidateNode, msgBytes); err != nil {
		n.logger.Printf("couldn't send prevote response to %v: %v\n", precandidateid, err.Error())
		return
	}

	if grant {
		n.logger.Printf("[DEBUG] raftify: Sent prevote response to %v (granted)\n", precandidateid)
	} else {
		n.logger.Printf("[DEBUG] raftify: Sent prevote response to %v (not granted)\n", precandidateid)
	}
}

// sendVoteRequest sends a vote request message to the nodes specified in the list passed in.
func (n *Node) sendVoteRequestToAll(list []*memberlist.Node) {
	reqBytes, _ := json.Marshal(VoteRequest{
		Term:        n.currentTerm,
		CandidateID: n.config.ID,
	})
	msgBytes, _ := json.Marshal(Message{
		Type:    VoteRequestMsg,
		Content: reqBytes,
	})

	for _, member := range list {
		if member.Name == n.config.ID {
			continue
		}
		if err := n.memberlist.SendBestEffort(member, msgBytes); err != nil {
			n.logger.Printf("[ERR] raftify: couldn't send vote request to %v: %v\n", member.Name, err.Error())
			continue
		}
		n.logger.Printf("[DEBUG] raftify: Sent vote request to %v\n", member.Name)
	}
}

// sendVoteResponse sends a vote response message back to the candidate who sent the vote request.
func (n *Node) sendVoteResponse(candidateid string, grant bool) {
	respBytes, _ := json.Marshal(VoteResponse{
		Term:        n.currentTerm,
		FollowerID:  n.config.ID,
		VoteGranted: grant,
	})
	msgBytes, _ := json.Marshal(Message{
		Type:    VoteResponseMsg,
		Content: respBytes,
	})

	candidateNode, err := n.getNodeByName(candidateid)
	if err != nil {
		n.logger.Printf("[ERR] raftify: %v\n", err.Error())
		return
	}

	if err := n.memberlist.SendBestEffort(candidateNode, msgBytes); err != nil {
		n.logger.Printf("[ERR] raftify: couldn't send vote response to %v: %v", candidateid, err.Error())
		return
	}

	if grant {
		n.logger.Printf("[DEBUG] raftify: Sent vote response to %v (granted)\n", candidateid)
	} else {
		n.logger.Printf("[DEBUG] raftify: Sent vote response to %v (not granted)\n", candidateid)
	}
}

// broadcastIntentionalLeave broadcasts an intentional leave message with the new quorum to all
// active members of the cluster.
func (n *Node) broadcastIntentionalLeave(newquorum int) {
	n.logger.Printf("[DEBUG] raftify: Enqueuing broadcast for quorum update (%v => %v) on intentional leave (active nodes: %v)", n.quorum, newquorum, n.memberlist.NumMembers())

	bcBytes, _ := json.Marshal(IntentionalLeave{
		NewQuorum: newquorum,
	})
	n.messages.broadcasts.QueueBroadcast(&Message{
		Content: bcBytes,
		Type:    IntentionalLeaveMsg,
	})
}
