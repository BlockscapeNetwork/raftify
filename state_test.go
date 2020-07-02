package raftify

import (
	"fmt"
	"os"
	"testing"
)

func TestSaveLoadDeleteState(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(3)

	// Initialize and start dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.config.PeerList = []string{
		fmt.Sprintf("127.0.0.1:%v", ports[1]),
		fmt.Sprintf("127.0.0.1:%v", ports[2]),
	}

	node.createMemberlist()
	defer node.memberlist.Shutdown()

	// Fail loading the current node state
	list, err := node.loadState()
	if err == nil {
		t.Log("Expected loadState to throw an error, instead it didn't")
		t.FailNow()
	}

	// Save the current node state
	node.saveState()
	if _, err := os.Stat(node.workingDir + "/state.json"); err != nil {
		t.Logf("Expected existing state.json, instead got error: %v", err.Error())
		t.FailNow()
	}

	// Succeed loading the current node state
	list, err = node.loadState()
	if err != nil {
		t.Logf("Expected state.json to be loaded successfully, instead got error: %v", err.Error())
		t.FailNow()
	}
	if len(list) != len(node.memberlist.Members()) {
		t.Logf("Expected loaded list to be equal in size to the internal memberlist, instead got sizes %v (loaded) and %v (internal)", len(list), len(node.memberlist.Members()))
		t.FailNow()
	}

	// Delete the current node state
	node.deleteState()
	if _, err := os.Stat(node.workingDir + "/state.json"); err == nil {
		t.Log("Expected no state.json, instead it exists")
		t.FailNow()
	}
}
