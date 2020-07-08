package raftify

import (
	"testing"
)

func TestGetNodeByName(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.createMemberlist()
	defer node.memberlist.Shutdown()

	if _, err := node.getNodeByName("TestNode"); err != nil {
		t.Logf("Expected TestNode to be found by its name, instead got error: %v", err.Error())
		t.FailNow()
	}
	if _, err := node.getNodeByName("NonExistentNode"); err == nil {
		t.Logf("Expected NonExistentNode to not be found, instead it was found")
		t.FailNow()
	}
}
