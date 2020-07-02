package raftify

import (
	"testing"
)

func TestToCandidate(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	// Check state transitions constraints
	node.state = Leader
	node.toCandidate()
	node.timeoutTimer.Stop()

	if node.state == Candidate {
		t.Log("Expected node to remain in the Leader state, instead it switched into the Candidate state")
		t.FailNow()
	}

	node.state = Follower
	node.toCandidate()
	node.timeoutTimer.Stop()

	if node.state == Candidate {
		t.Log("Expected node to remain in the Follower state, instead it switched into the Candidate state")
		t.FailNow()
	}

	// Switch into Candidate state
	node.state = PreCandidate
	node.toCandidate()
	node.timeoutTimer.Stop()

	if node.state != Candidate {
		t.Logf("Expected node to be in the Candidate state, instead got %v", node.state.toString())
		t.FailNow()
	}
	if node.currentTerm != 1 {
		t.Logf("Expected node to be at term 1, instead got %v", node.currentTerm)
		t.FailNow()
	}
	if node.votedFor != node.config.ID {
		t.Logf("Expected node to have voted for itself, instead got %v", node.votedFor)
		t.FailNow()
	}
}
