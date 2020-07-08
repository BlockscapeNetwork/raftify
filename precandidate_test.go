package raftify

import (
	"testing"
)

func TestToPreCandidate(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()

	// Switch into PreCandidate state
	node.toPreCandidate()
	if node.state != PreCandidate {
		t.Logf("Expected node to be in the PreCandidate state, instead got %v", node.state.toString())
		t.FailNow()
	}
}

func TestRunPreCandidateTimeoutElapsedCase(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.quorum = 1

	node.createMemberlist()
	defer node.memberlist.Shutdown()

	done := make(chan bool)

	// Test quorum reached for single node cluster case
	go func() {
		node.runPreCandidate()
		done <- true
	}()

	node.toPreCandidate()
	<-done

	if node.state != Candidate {
		t.Logf("Expected node to be in the Candidate state, instead got %v", node.state.toString())
		t.FailNow()
	}

	// Test quorum reached and missed prevote cycles
	go func() {
		node.runPreCandidate()
		done <- true
	}()

	node.quorum = 2
	node.preVoteList.missedPrevoteCycles = 4

	node.toPreCandidate()
	<-done

	if node.preVoteList.missedPrevoteCycles != 0 {
		t.Logf("Expected missedPrevoteCycles to be 0, instead got %v", node.preVoteList.missedPrevoteCycles)
		t.FailNow()
	}
	if !node.rejoin {
		t.Logf("Expected the rejoin flag to be set to true, instead got %v", node.rejoin)
		t.FailNow()
	}
	if node.state != Follower {
		t.Logf("Expected node to be in the Follower state, instead got %v", node.state.toString())
		t.FailNow()
	}
}
