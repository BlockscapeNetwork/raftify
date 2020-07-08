package raftify

import (
	"fmt"
	"testing"
	"time"
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

func TestRunFollowerTimeoutElapsedCase(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(2)

	// Initialize dummy nodes
	node1 := initDummyNode("TestNode_1", 1, 2, ports[0])
	node2 := initDummyNode("TestNode_2", 2, 2, ports[1])

	// Start node1 for the first test with the election timeout
	node1.createMemberlist()
	defer node1.memberlist.Shutdown()

	time.Sleep(3 * time.Second)

	// Wait for election timeout to elapse and the node to switch to the PreCandidate state
	node1.resetTimeout()
	node1.runFollower()
	node1.timeoutTimer.Stop()

	if node1.state != PreCandidate {
		t.Logf("Expected node1 to be in the PreCandidate state, instead got %v", node1.state.toString())
		t.FailNow()
	}

	// Wait for the election timeout to elapse and the node to initiate a rejoin
	node1.rejoin = true
	node1.config.PeerList = []string{fmt.Sprintf("127.0.0.1:%v", node2.config.BindPort)}

	node1.resetTimeout()
	node1.runFollower() // Returns because node1 can't join node2
	node1.timeoutTimer.Stop()

	// Start node2 for the second test with the rejoin and let it rejoin node1
	node2.createMemberlist()
	defer node2.memberlist.Shutdown()

	node2.rejoin = true
	node2.config.PeerList = []string{fmt.Sprintf("127.0.0.1:%v", node1.config.BindPort)}

	node2.resetTimeout()
	go node2.runFollower() // Joins node1 and resets rejoin flag
	<-node2.bootstrapCh
	node2.timeoutTimer.Stop()

	if node2.rejoin {
		t.Logf("Expected rejoin flag to be false, instead got %v", node2.rejoin)
		t.FailNow()
	}
}
