package raftify

import (
	"fmt"
	"testing"
	"time"
)

func TestTryJoin(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(2)

	// Initialize dummy nodes
	node1 := initDummyNode("TestNode_1", 1, 2, ports[0])
	node2 := initDummyNode("TestNode_2", 1, 2, ports[1])

	node1.config.PeerList = []string{fmt.Sprintf("127.0.0.1:%v", node2.config.BindPort)}
	node2.config.PeerList = []string{fmt.Sprintf("127.0.0.1:%v", node1.config.BindPort)}

	// Start node1 and fail while trying to join node2
	node1.createMemberlist()
	defer node1.memberlist.Shutdown()

	if err := node1.tryJoin(); err == nil {
		t.Logf("Expected node1 to throw an error on tryJoin, instead error was nil")
		t.FailNow()
	}

	// Start node2 and succeed while trying to join node1
	node2.createMemberlist()
	defer node2.memberlist.Shutdown()

	if err := node2.tryJoin(); err != nil {
		t.Logf("Expected node2 to successfully join node1, instead got error: %v", err.Error())
		t.FailNow()
	}
}

func TestToBootstrap(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(2)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 2, 2, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	// Switch into Bootstrap state, expecting another node to join
	node.toBootstrap()
	if node.state != Bootstrap {
		t.Logf("Expected node to be in the Bootstrap state, instead got %v", node.state.toString())
		t.FailNow()
	}

	// Switch into Bootstrap state, expecting a single node cluster with no peers
	node.config.Expect = 1
	node.toBootstrap()

	if node.state != Leader {
		t.Logf("Expected node to be in the Leader state, instead got %v", node.state.toString())
		t.FailNow()
	}

	// Switch into Bootstrap state, expecting a single node cluster with peers
	node.config.PeerList = []string{fmt.Sprintf("127.0.0.1:%v", ports[1])}
	node.toBootstrap()

	if node.state != Follower {
		t.Logf("Expected node to be in the Follower state, instead got %v", node.state.toString())
		t.FailNow()
	}
}

func TestRunBootstrapJoinEventCase(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(2)

	// Initialize dummy nodes
	node1 := initDummyNode("TestNode_1", 2, 2, ports[0])
	node2 := initDummyNode("TestNode_2", 2, 2, ports[1])

	node1.config.PeerList = []string{fmt.Sprintf("127.0.0.1:%v", node2.config.BindPort)}
	node2.config.PeerList = []string{fmt.Sprintf("127.0.0.1:%v", node1.config.BindPort)}

	// Start both nodes
	node1.createMemberlist()
	node2.createMemberlist()
	defer node1.memberlist.Shutdown()
	defer node2.memberlist.Shutdown()

	// Run runBootstrap in goroutine, so the event case gets triggered on join
	go node2.runBootstrap()

	// Form a cluster
	node2.tryJoin()

	select {
	case <-node2.bootstrapCh:
		if node2.state != Follower {
			t.Logf("Expected node2 to be in the Follower state, instead got %v", node2.state.toString())
			t.FailNow()
		}
		break
	case <-time.After(6 * time.Second):
		t.Log("Expected join event to happen, instead nothing happened")
		t.FailNow()
	}
}
