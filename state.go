package raftify

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/hashicorp/memberlist"
)

// saveState saves the current memberlist into a separate state.json file. This file is used
// to allow a timed out or crashed node which has lost its internal memberlist to rejoin the
// cluster it is already part of. The state.json file is generated on the first successful join.
func (n *Node) saveState() error {
	stateJSON, _ := json.MarshalIndent(n.memberlist.Members(), "", "	")
	_ = ioutil.WriteFile(n.workingDir+"/state.json", stateJSON, 0755)
	n.logger.Println("[DEBUG] raftify: Created/Updated state.json ✓")
	return nil
}

// deleteState deletes the state.json. Called only when the node explicitly leaves on its own
// accord.
func (n *Node) deleteState() error {
	os.Remove(n.workingDir + "/state.json")
	n.logger.Println("[DEBUG] raftify: Deleted state.json ✓")
	return nil
}

// loadState loads the contents of the state.json file.
func (n *Node) loadState() ([]*memberlist.Node, error) {
	stateJSON, err := os.Open(n.workingDir + "/state.json")
	if err != nil {
		return nil, err
	}
	defer stateJSON.Close()

	stateBytes, err := ioutil.ReadAll(stateJSON)
	if err != nil {
		return nil, err
	}

	var list []*memberlist.Node
	if err = json.Unmarshal(stateBytes, &list); err != nil {
		return nil, err
	}
	return list, nil
}
