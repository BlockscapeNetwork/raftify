package raftify

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestAPI(t *testing.T) {
	// Setup configuration data
	config := Config{
		ID:       "Node_TestAPI",
		MaxNodes: 1,
		Expect:   1,
		LogLevel: "DEBUG",
		BindAddr: "127.0.0.1",
		BindPort: 3000,
	}
	pwd, _ := os.Getwd()
	logger := log.New(os.Stderr, "", 0)

	// Create directory for test data
	os.MkdirAll(pwd+"/testing/TestNode", 0755)
	defer os.RemoveAll(pwd + "/testing")

	// Write configuration data to raftify.json file
	nodesBytes, _ := json.Marshal(config)
	ioutil.WriteFile(pwd+"/testing/TestNode/raftify.json", nodesBytes, 0755)

	// Test InitNode
	node, err := InitNode(logger, pwd+"/testing/TestNode")
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
	if _, ok := members["Node_TestAPI"]; !ok {
		t.Log("Expected to find member \"Node_TestAPI\", instead not found")
		t.FailNow()
	}
	if len(members) != 1 {
		t.Logf("Expected length of memberlist to be 1, instead got %v", len(members))
		t.FailNow()
	}

	// Test GetID
	if node.GetID() != "Node_TestAPI" {
		t.Logf("Expected id to be \"Node_TestAPI\", instead got %v", members["id"])
		t.FailNow()
	}

	// Test GetState
	if node.GetState() != Leader {
		t.Logf("Expected node to be leader, instead got %v", node.state.toString())
		t.FailNow()
	}

	// Test Shutdown
	if err := node.Shutdown(); err != nil {
		t.Logf("Expected successful shutdown of Node_TestAPI, instead got error: %v", err.Error())
		t.FailNow()
	}
}
