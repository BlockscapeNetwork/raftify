package raftify

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/memberlist"
)

// Message is a wrapper struct for all messages used to determine the message type.
type Message struct {
	Type    MessageType     `json:"type"`
	Content json.RawMessage `json:"content"`
}

// Heartbeat defines the message sent out by the leader to all cluster members.
type Heartbeat struct {
	Term        uint64 `json:"term"`
	Quorum      int    `json:"quorum"`
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

// NewQuorum defines the message sent out by a node that is voluntarily leaving the cluster,
// triggering an immediate quorum change. This does not include crash-related leave events.
type NewQuorum struct {
	NewQuorum int `json:"new_quorum"`
}

// sendHeartbeatToAll sends a heartbeat message to all the other cluster members.
func (n *Node) sendHeartbeatToAll() {
	n.heartbeatIDList.reset()

	hb := Heartbeat{
		HeartbeatID: n.heartbeatIDList.currentHeartbeatID,
		Term:        n.currentTerm,
		Quorum:      n.quorum,
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
func (n *Node) sendPreVoteRequestToAll() {
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

// sendNewQuorumToAll sends the new quorum to the rest of the cluster triggered by a voluntary
// leave event. Once memberlist has processed the leave event internally, this message is used
// to trigger an immediate change of the new quorum instead of waiting for the dead node to
// be kicked. This function returns the number of nodes that the new quorum could be sent to.
func (n *Node) sendNewQuorumToAll(newquorum int) int {
	nqBytes, _ := json.Marshal(NewQuorum{
		NewQuorum: newquorum,
	})
	msgBytes, _ := json.Marshal(Message{
		Type:    NewQuorumMsg,
		Content: nqBytes,
	})

	// Count how many members received the new quorum message
	membersReached := 0

	for _, member := range n.memberlist.Members() {
		if member.Name == n.config.ID {
			continue
		}
		if err := n.memberlist.SendReliable(member, msgBytes); err != nil {
			n.logger.Printf("[ERR] raftify: couldn't send new quorum to %v: %v\n", member.Name, err.Error())
			continue
		}
		membersReached++
	}
	return membersReached
}
