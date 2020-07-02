package raftify

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

func TestAPI(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])

	// Create directory for test data
	os.MkdirAll(node.workingDir+"/testing/TestNode", 0755)
	defer os.RemoveAll(node.workingDir + "/testing")

	// Write configuration data to raftify.json file
	nodesBytes, _ := json.Marshal(node.config)
	ioutil.WriteFile(node.workingDir+"/testing/TestNode/raftify.json", nodesBytes, 0755)

	// Test InitNode
	node, err := InitNode(node.logger, node.workingDir+"/testing/TestNode")
	if err != nil {
		t.Logf("Expected node to initialize successfully, instead got error: %v", err.Error())
		t.FailNow()
	}

	// Test GetHealthScore
	if node.GetHealthScore() != 0 {
		t.Logf("Expected node to reach a health score of 0, instead got %v", node.GetHealthScore())
		t.FailNow()
	}

	// Test GetMembers
	members := node.GetMembers()
	if _, ok := members["TestNode"]; !ok {
		t.Logf("Expected to find member \"%v\", instead not found", node.config.ID)
		t.FailNow()
	}
	if len(members) != 1 {
		t.Logf("Expected length of memberlist to be 1, instead got %v", len(members))
		t.FailNow()
	}

	// Test GetID
	if node.GetID() != "TestNode" {
		t.Logf("Expected id to be \"%v\", instead got %v", node.config.ID, members["id"])
		t.FailNow()
	}

	// Test GetState
	if node.GetState() != Leader {
		t.Logf("Expected node to be leader, instead got %v", node.state.toString())
		t.FailNow()
	}

	// Test Shutdown
	if err := node.Shutdown(); err != nil {
		t.Logf("Expected successful shutdown of %v, instead got error: %v", node.config.ID, err.Error())
		t.FailNow()
	}
}
