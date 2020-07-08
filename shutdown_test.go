package raftify

import (
	"testing"
	"time"
)

func TestToShutdown(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	// Switch into Shutdown state
	node.toShutdown()
	if node.state != Shutdown {
		t.Logf("Expected node to be in the Shutdown state, instead got %v", node.state.toString())
		t.FailNow()
	}
}

func TestRunShutdown(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	done := make(chan bool)

	// Wait for runShutdown to finish
	go func() {
		<-node.shutdownCh
		done <- true
	}()

	node.runShutdown()

	select {
	case <-done:
		break
	case <-time.After(3 * time.Second):
		t.Log("Expected node to shut down, instead nothing happened")
		t.FailNow()
	}
}
