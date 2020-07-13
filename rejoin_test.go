package raftify

import (
	"fmt"
	"testing"
)

func TestToRejoin(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	// Switch into Rejoin state
	node.toRejoin(false)
	if node.state != Rejoin {
		t.Logf("Expected node to be in the Rejoin state, instead got %v", node.state.toString())
		t.FailNow()
	}
}

func TestRunRejoin(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(2)

	// Initialize and start dummy node
	node1 := initDummyNode("TestNode_1", 1, 2, ports[0])
	node2 := initDummyNode("TestNode_2", 2, 2, ports[1])

	node1.createMemberlist()
	defer node1.memberlist.Shutdown()

	// Wait for the election timeout to elapse and the node to initiate a rejoin
	node1.toRejoin(false)
	node1.config.PeerList = []string{fmt.Sprintf("127.0.0.1:%v", node2.config.BindPort)}

	node1.resetTimeout()
	node1.runRejoin() // Returns because node1 can't join node2
	node1.timeoutTimer.Stop()

	// Start node2 for the second test with the rejoin and let it rejoin node1
	node2.createMemberlist()
	defer node2.memberlist.Shutdown()

	node2.toRejoin(true)
	node2.config.PeerList = []string{fmt.Sprintf("127.0.0.1:%v", node1.config.BindPort)}

	// Catch signal on bootstrap channel as node2 is in Initialize state
	go func() {
		<-node2.bootstrapCh
	}()

	node2.resetTimeout()
	node2.runRejoin() // Joins node1 and switches into the Follower state
	node2.timeoutTimer.Stop()

	if node2.state != Follower {
		t.Logf("Expected node to be in the Follower state, instead got: %v", node2.state.toString())
		t.FailNow()
	}
}
