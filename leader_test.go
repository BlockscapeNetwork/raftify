package raftify

import (
	"testing"
	"time"
)

func TestToLeader(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()

	// Check state transitions constraints
	node.toFollower(0)
	node.toLeader()
	node.messageTicker.Stop()

	if node.state == Leader {
		t.Log("Expected node to remain in the Follower state, instead it switched into the Leader state")
		t.FailNow()
	}

	// Switch into Leader state
	node.toFollower(0)
	node.toPreCandidate()
	node.toCandidate()
	node.toLeader()
	node.messageTicker.Stop()

	if node.state != Leader {
		t.Logf("Expected node to be in the Leader state, instead got %v", node.state.toString())
		t.FailNow()
	}
	if node.votedFor != "" {
		t.Logf("Expected node to not have voted for anyone, instead got %v", node.votedFor)
		t.FailNow()
	}

	// Shutdown node
	node.memberlist.Shutdown()
}

func TestRunLeaderTickerCase(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.quorum = 1
	node.heartbeatIDList.received = 1
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	done := make(chan bool)

	// Check sub quorum cycles after having reached the quorum
	go func() {
		node.runLeader()
		done <- true
	}()

	node.messageTicker = time.NewTicker(10 * time.Millisecond)
	<-done

	if node.heartbeatIDList.subQuorumCycles != 0 {
		t.Logf("Expected subQuorumCycles to be reset to 0, instead got %v", node.heartbeatIDList.subQuorumCycles)
		t.FailNow()
	}

	node.messageTicker.Stop()

	// Check sub quorum cycles after failing to reach the quorum
	node.heartbeatIDList.received = 0

	go func() {
		node.runLeader()
		done <- true
	}()

	node.messageTicker = time.NewTicker(10 * time.Millisecond)
	<-done

	if node.heartbeatIDList.subQuorumCycles != 1 {
		t.Logf("Expected subQuorumCycles to be 1, instead got %v", node.heartbeatIDList.subQuorumCycles)
		t.FailNow()
	}

	node.messageTicker.Stop()

	// Check if maximum sub quorum cycles have been reached
	node.heartbeatIDList.received = 0
	node.heartbeatIDList.subQuorumCycles = MaxSubQuorumCycles

	go func() {
		node.runLeader()
		done <- true
	}()

	node.messageTicker = time.NewTicker(10 * time.Millisecond)
	<-done

	if !node.rejoin {
		t.Logf("Expected rejoin flag to be set to true, instead got %v", node.rejoin)
		t.FailNow()
	}
	if node.heartbeatIDList.currentHeartbeatID != 0 {
		t.Logf("Expected current heartbeat id to be 0, instead got %v", node.heartbeatIDList.currentHeartbeatID)
		t.FailNow()
	}
	if node.heartbeatIDList.subQuorumCycles != 0 {
		t.Logf("Expected subQuorumCycles to be reset to 0, instead got %v", node.heartbeatIDList.subQuorumCycles)
		t.FailNow()
	}
	if node.state != Follower {
		t.Logf("Expected node to be in the Follower state, instead got %v", node.state.toString())
		t.FailNow()
	}
}
