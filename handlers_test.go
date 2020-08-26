package raftify

import (
	"testing"
	"time"

	"github.com/hashicorp/memberlist"
)

func TestHandleHeartbeatAsFollower(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	// Switch into Follower state for term 0
	node.toFollower(0)

	// Make heartbeat message
	hb := Heartbeat{
		LeaderID: "TestNode",
	}

	// Higher term arrives
	hb.Term = node.currentTerm + 1
	node.handleHeartbeat(hb)

	if node.currentTerm != hb.Term {
		t.Logf("Expected node to adopt term %v, instead got %v", hb.Term, node.currentTerm)
		t.FailNow()
	}

	// Lower term arrives
	hb.Term = node.currentTerm - 1
	node.handleHeartbeat(hb)

	if node.currentTerm == hb.Term {
		t.Logf("Expected node to keep its current term, instead it adopted %v", hb.Term)
		t.FailNow()
	}

	// Same term arrives
	hb.Term = node.currentTerm
	node.handleHeartbeat(hb)

	select {
	case <-node.timeoutTimer.C:
		break
	case <-time.After((MaxTimeout*time.Duration(node.config.Performance) + 1) * time.Millisecond):
		t.Logf("Expected election timeout to be reset after receival of heartbeat with same term, instead nothing happened")
		t.FailNow()
	}
}

func TestHandleHeartbeatAsPreCandidate(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	// Make heartbeat message
	hb := Heartbeat{
		LeaderID: "TestNode",
	}

	// Higher term arrives
	node.toPreCandidate()
	hb.Term = node.currentTerm + 1
	node.handleHeartbeat(hb)

	if node.state != Follower {
		t.Logf("Expected node to be in the Follower state, instead got %v", node.state.toString())
		t.FailNow()
	}
	if node.currentTerm != hb.Term {
		t.Logf("Expected node to adopt term %v, instead got %v", hb.Term, node.currentTerm)
		t.FailNow()
	}

	// Same or lower term
	node.toPreCandidate()
	hb.Term = node.currentTerm - 1
	node.handleHeartbeat(hb)

	if node.state != PreCandidate {
		t.Logf("Expected node to be in the PreCandidate state, instead got %v", node.state.toString())
		t.FailNow()
	}
	if node.currentTerm != hb.Term+1 {
		t.Logf("Expected node to keep its term, instead it adopted term %v", hb.Term)
		t.FailNow()
	}
}

func TestHandleHeartbeatAsCandidate(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	// Make heartbeat message
	hb := Heartbeat{
		LeaderID: "TestNode",
	}

	// Same or higher term arrives
	node.toCandidate()
	hb.Term = node.currentTerm + 1
	node.handleHeartbeat(hb)

	if node.state != Follower {
		t.Logf("Expected node to be in the Follower state, instead got %v", node.state.toString())
		t.FailNow()
	}
	if node.currentTerm != hb.Term {
		t.Logf("Expected node to have adopted term %v, instead it didn't", hb.Term)
		t.FailNow()
	}

	// Lower term
	node.toPreCandidate()
	node.toCandidate()
	hb.Term = node.currentTerm - 1
	node.handleHeartbeat(hb)

	if node.state != Candidate {
		t.Logf("Expected node to be in the Candidate state, instead got %v", node.state.toString())
		t.FailNow()
	}
	if node.currentTerm != hb.Term+1 {
		t.Logf("Expected node to have adopted term %v, instead it didn't", hb.Term)
		t.FailNow()
	}
}

func TestHandleHeartbeatAsLeader(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	// Make heartbeat message
	hb := Heartbeat{
		LeaderID: "TestNode",
	}

	// Catch the panic
	defer func(t *testing.T) {
		if r := recover(); r == nil {
			t.Logf("Expected code to panic if leader receives a heartbeat, instead it did not")
			t.FailNow()
		}
	}(t)

	node.toLeader()
	node.handleHeartbeat(hb)
}

func TestHandleHeartbeatResponse(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	// Make heartbeat response message
	hb := HeartbeatResponse{
		FollowerID: "TestNode",
	}

	// Invalid state
	node.toFollower(0)
	node.handleHeartbeatResponse(hb)

	if node.heartbeatIDList.received != 0 {
		t.Logf("Expected to not have received any heartbeat IDs, instead got %v", node.heartbeatIDList.received)
		t.FailNow()
	}

	// Higher term arrives
	node.toPreCandidate()
	node.toCandidate()
	node.toLeader()

	hb.Term = node.currentTerm + 1
	node.handleHeartbeatResponse(hb)

	if node.state != Follower {
		t.Logf("Expected node to be in the Follower state, instead got %v", node.state.toString())
		t.FailNow()
	}
	if node.heartbeatIDList.received != 1 {
		t.Logf("Expected to not have received any heartbeat IDs, instead got %v", node.heartbeatIDList.received)
		t.FailNow()
	}
	if node.currentTerm != hb.Term {
		t.Logf("Expected node to have adopted term %v, instead it didn't", hb.Term)
		t.FailNow()
	}

	// Lower term arrives
	hb.Term = node.currentTerm - 1
	node.handleHeartbeatResponse(hb)

	if node.heartbeatIDList.received != 1 {
		t.Logf("Expected to not have received any heartbeat IDs, instead got %v", node.heartbeatIDList.received)
		t.FailNow()
	}
	if node.currentTerm != hb.Term+1 {
		t.Logf("Expected node to have adopted term %v, instead it didn't", hb.Term)
		t.FailNow()
	}

	// Same term arrives
	hb.Term = node.currentTerm
	node.handleHeartbeatResponse(hb)

	if node.heartbeatIDList.received == 0 {
		t.Logf("Expected node to have received one heartbeat ID, instead got %v", node.heartbeatIDList.received)
		t.FailNow()
	}
}

func ExampleNode_handlePreVoteRequest() {
	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, 5000)
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	// Make prevote request message
	pvr := PreVoteRequest{
		PreCandidateID: "TestNode",
	}

	// Invalid state, not granted
	node.toFollower(0)
	node.handlePreVoteRequest(pvr)

	// Valid state, not granted
	node.toPreCandidate()
	node.handlePreVoteRequest(pvr)

	// Valid state, granted
	pvr.NextTerm = node.currentTerm + 1
	node.handlePreVoteRequest(pvr)
	// Output:
	// [INFO] raftify: ->[] TestNode [127.0.0.1:5000] joined the cluster.
	// [INFO] raftify: Entering follower state for term 0
	// [DEBUG] raftify: Received prevote request from TestNode
	// [WARN] raftify: received prevote request as Follower
	// [DEBUG] raftify: Sent prevote response to TestNode (not granted)
	// [DEBUG] raftify: Entering precandidate state for term 1
	// [DEBUG] raftify: Received prevote request from TestNode
	// [DEBUG] raftify: Received outdated prevote request from TestNode, skipping...
	// [DEBUG] raftify: Sent prevote response to TestNode (not granted)
	// [DEBUG] raftify: Received prevote request from TestNode
	// [DEBUG] raftify: Sent prevote response to TestNode (granted)
}

func TestHandlePreVoteResponse(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	// Make prevote response message
	pvr := PreVoteResponse{
		FollowerID: "TestNode",
	}

	// Invalid state
	node.toFollower(0)
	node.handlePreVoteResponse(pvr)

	if node.state != Follower {
		t.Logf("Expected node to be in the Follower state, instead got %v", node.state.toString())
		t.FailNow()
	}

	// Valid state, higher term arrives
	node.toPreCandidate()
	pvr.Term = node.currentTerm + 1
	node.handlePreVoteResponse(pvr)

	if node.state != Follower {
		t.Logf("Expected node to be in the Follower state, instead got %v", node.state.toString())
		t.FailNow()
	}
	if node.currentTerm != pvr.Term {
		t.Logf("Expected node to have adopted term %v, instead it didn't", pvr.Term)
		t.FailNow()
	}

	// Valid state, lower term arrives
	pvr.Term = node.currentTerm - 1
	node.handlePreVoteResponse(pvr)

	if node.state != Follower {
		t.Logf("Expected node to be in the Follower state, instead got %v", node.state.toString())
		t.FailNow()
	}
	if node.currentTerm != pvr.Term+1 {
		t.Logf("Expected node to have adopted term %v, instead it didn't", pvr.Term)
		t.FailNow()
	}

	// Valid state, same term
	node.toPreCandidate()
	pvr.Term = node.currentTerm
	node.handlePreVoteResponse(pvr)

	if node.preVoteList.missedPrevoteCycles != 0 {
		t.Logf("Expected missed prevote cycles to be 0, instead got %v", node.preVoteList.missedPrevoteCycles)
		t.FailNow()
	}

	if node.preVoteList.received != 1 {
		t.Logf("Expected node to have received one heartbeat ID, instead got %v", node.preVoteList.received)
		t.FailNow()
	}

	// Prevote granted
	node.quorum = 1
	pvr.PreVoteGranted = true
	node.preVoteList.pending = append(node.preVoteList.pending, node.memberlist.LocalNode())
	node.handlePreVoteResponse(pvr)

	if node.preVoteList.received != 2 {
		t.Logf("Expected node to have received two heartbeat IDs, instead got %v", node.preVoteList.received)
		t.FailNow()
	}
	if node.state != Candidate {
		t.Logf("Expected node to be in the Candidate state, instead got %v", node.state.toString())
		t.FailNow()
	}
}

func TestHandleVoteRequest(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	vr := VoteRequest{
		CandidateID: "TestNode",
	}

	// Higher term arrives
	vr.Term = node.currentTerm + 1
	node.handleVoteRequest(vr)

	if node.state != Follower {
		t.Logf("Expected node to be in the Follower state, instead got %v", node.state.toString())
		t.FailNow()
	}
	if node.currentTerm != vr.Term {
		t.Logf("Expected node to have adopted term %v, instead it didn't", vr.Term)
		t.FailNow()
	}

	// Lower term arrives
	node.toPreCandidate()
	node.toCandidate()

	vr.Term = node.currentTerm - 1
	node.handleVoteRequest(vr)

	if node.state != Candidate {
		t.Logf("Expected node to be in the Candidate state, instead got %v", node.state.toString())
		t.FailNow()
	}
}

func TestHandleVoteResponse(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	vr := VoteResponse{
		FollowerID: "TestNode",
	}

	// Invalid state
	node.toFollower(0)
	node.handleVoteResponse(vr)

	if node.state != Follower {
		t.Logf("Expected node to be in the Follower state, instead got %v", node.state.toString())
		t.FailNow()
	}

	// Valid state, same term arrives
	node.toPreCandidate()
	node.toCandidate()

	vr.Term = node.currentTerm
	node.voteList.pending = append(node.voteList.pending, node.memberlist.LocalNode())
	node.handleVoteResponse(vr)

	if node.voteList.received != 1 {
		t.Logf("Expected node to have received one vote response, instead got %v", node.voteList.received)
		t.FailNow()
	}

	node.quorum = 1
	vr.VoteGranted = true
	node.voteList.pending = append(node.voteList.pending, node.memberlist.LocalNode())
	node.handleVoteResponse(vr)

	if node.state != Leader {
		t.Logf("Expected node to be in the Leader state, instead got %v", node.state.toString())
		t.FailNow()
	}
}

func TestHandleNewQuorum(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	nq := NewQuorum{
		NewQuorum: 2,
		LeavingID: "TestNode",
	}

	done := make(chan bool)
	defer node.deleteState()

	// Test case if new quorum greater than 1 is handled and leave event is fired
	go func() {
		node.handleNewQuorum(nq)
		done <- true
	}()

	node.events.eventCh <- memberlist.NodeEvent{
		Event: memberlist.NodeLeave,
	}
	<-done

	if node.quorum != 2 {
		t.Logf("Expected the quorum to be 2, instead got %v", node.quorum)
		t.FailNow()
	}
	if node.state == Leader {
		t.Logf("Expected node to be in any other state but the Leader state, instead got %v", node.state.toString())
		t.FailNow()
	}

	// Test case if new quorum is 1 and leave event is fired
	nq.NewQuorum = 1

	go func() {
		node.handleNewQuorum(nq)
		done <- true
	}()

	node.events.eventCh <- memberlist.NodeEvent{
		Event: memberlist.NodeLeave,
	}
	<-done

	if node.quorum != 1 {
		t.Logf("Expected the quorum to be 1, instead got %v", node.quorum)
		t.FailNow()
	}
	if node.state != Leader {
		t.Logf("Expected node to be in the Leader state, instead got %v", node.state.toString())
		t.FailNow()
	}

	// Test case if join event is fired
	nq.NewQuorum = 0

	go func() {
		node.handleNewQuorum(nq)
		done <- true
	}()

	node.events.eventCh <- memberlist.NodeEvent{
		Event: memberlist.NodeJoin,
	}

	select {
	case <-time.After(3 * time.Second):
		break
	case <-done:
		t.Logf("Expected node to keep waiting on a leave event, instead it returned on a join event")
		t.FailNow()
	}

	if node.quorum != 1 {
		t.Logf("Expected quorum to have stayed 1, instead got %v", node.quorum)
		t.FailNow()
	}
}
