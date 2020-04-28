package raftify

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestAPI(t *testing.T) {
	config := Config{
		ID:       "Node_TestAPI",
		MaxNodes: 1,
		Expect:   1,
		LogLevel: "DEBUG",
		BindPort: 3000,
	}
	pwd, _ := os.Getwd()
	logger := log.New(os.Stderr, "", 0)

	os.MkdirAll(pwd+"/testing/Node-0", 0755)
	defer os.RemoveAll(pwd + "/testing")

	nodesBytes, _ := json.Marshal(config)
	ioutil.WriteFile(pwd+"/testing/Node-0/raftify.json", nodesBytes, 0755)

	node, err := InitNode(logger, pwd+"/testing/Node-0")
	if err != nil {
		t.Logf("Expected node to initialize successfully, instead got error: %v", err.Error())
		t.FailNow()
	}

	if node.GetHealthScore() != 0 {
		t.Logf("Expected node to reach a health score of 0, instead got %v", node.GetHealthScore())
		t.FailNow()
	}

	members := node.GetMembers()
	if _, ok := members["Node_TestAPI"]; !ok {
		t.Log("Expected to find member \"Node_TestAPI\", instead not found")
		t.FailNow()
	}
	if len(members) != 1 {
		t.Logf("Expected length of memberlist to be 1, instead got %v", len(members))
		t.FailNow()
	}

	if node.GetID() != "Node_TestAPI" {
		t.Logf("Expected id to be \"Node_TestAPI\", instead got %v", members["id"])
		t.FailNow()
	}

	if node.GetState() != Leader {
		t.Logf("Expected node to be leader, instead got %v", node.state.toString())
		t.FailNow()
	}

	if err := node.Shutdown(); err != nil {
		t.Logf("Expected successful shutdown of Node_TestAPI, instead got error: %v", err.Error())
		t.FailNow()
	}
}
