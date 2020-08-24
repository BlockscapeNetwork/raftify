package raftify

import (
	"fmt"
	"testing"
)

func TestToPreShutdown(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	// Switch into Shutdown state
	node.toPreShutdown()
	if node.state != PreShutdown {
		t.Logf("Expected node to be in the PreShutdown state, instead got %v", node.state.toString())
		t.FailNow()
	}
}

func TestRunPreShutdown(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(2)

	// Initialize and start dummy node
	node1 := initDummyNode("TestNode_1", 2, 2, ports[0])
	node2 := initDummyNode("TestNode_2", 2, 2, ports[1])

	peerlist := []string{
		fmt.Sprintf("127.0.0.1:%v", ports[0]),
		fmt.Sprintf("127.0.0.1:%v", ports[1]),
	}

	node1.createMemberlist()
	node2.createMemberlist()

	defer node1.memberlist.Shutdown()
	defer node2.memberlist.Shutdown()

	// Form cluster
	node1.memberlist.Join(peerlist)

	// Run preshotdown with cluster size greater than 1 (2 in this case)
	node1.runPreShutdown()
	if node1.state != Shutdown {
		t.Logf("Expected node1 to be in the Shutdown state, instead got %v", node1.state.toString())
		t.FailNow()
	}

	// Run preshutdown with cluster size of 1
	node2.memberlist.Shutdown()
	node1.runPreShutdown()
	if node1.state != Shutdown {
		t.Logf("Expected node1 to be in the Shutdown state, instead got %v", node1.state.toString())
		t.FailNow()
	}
}
