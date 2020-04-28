package raftify

import (
	"log"
	"os"
	"testing"

	"github.com/hashicorp/memberlist"
)

func TestSaveLoadDeleteState(t *testing.T) {
	pwd, _ := os.Getwd()
	config := Config{
		ID:       "Node_TestSaveLoadDeleteState",
		MaxNodes: 3,
		Expect:   1,
		BindAddr: "0.0.0.0",
		BindPort: 3000,
		PeerList: []string{
			"0.0.0.0:3001",
			"0.0.0.0:3002",
		},
	}
	node := &Node{
		logger:     log.New(os.Stderr, "", 0),
		workingDir: pwd,
		config:     &config,
	}

	var err error
	if node.memberlist, err = memberlist.Create(memberlist.DefaultWANConfig()); err != nil {
		t.Logf("Expected creation of memberlist, instead got error: %v", err.Error())
		t.Fail()
	}

	node.saveState()
	if _, err := os.Stat(pwd + "/state.json"); err != nil {
		t.Logf("Expected existing state.json, instead got error: %v", err.Error())
		t.Fail()
	}

	list, err := node.loadState()
	if err != nil {
		t.Logf("Expected no errors during loadState, instead got: %v", err.Error())
		t.Fail()
	}
	if len(list) != len(node.memberlist.Members()) {
		t.Logf("Expected loaded list to be equal in size to the internal memberlist, instead got %v (loaded) and %v (internal)", len(list), len(node.memberlist.Members()))
		t.Fail()
	}

	node.deleteState()
	if _, err := os.Stat(pwd + "/state.json"); err == nil {
		t.Log("Expected no state.json, instead it exists")
		t.Fail()
	}

	node.memberlist.Leave(0)
	node.memberlist.Shutdown()
}
