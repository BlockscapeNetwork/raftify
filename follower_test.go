package raftify

import (
	"testing"
)

func TestToFollower(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	// Switch into Follower state
	node.toFollower(0)
	if node.state != Follower {
		t.Logf("Expected node to be in the Bootstrap state, instead got %v", node.state.toString())
		t.FailNow()
	}
	if node.currentTerm != 0 {
		t.Logf("Expected node to be at term 0, instead got %v", node.currentTerm)
		t.FailNow()
	}
	if node.votedFor != "" {
		t.Logf("Expected node to not have voted for anyone, instead got %v", node.votedFor)
		t.FailNow()
	}
}
