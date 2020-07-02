package raftify

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"
)

func TestSingleNodeClusterWithNoPeers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestSingleNodeClusterWithNoPeers in short mode")
	}

	// Reserve ports for this test
	ports := reservePorts(1)

	config := Config{
		ID:       "Node_TestSingleNodeClusterWithNoPeers",
		MaxNodes: 1,
		Expect:   1,
		BindPort: ports[0],
	}

	// Initialize node.
	pwd, _ := os.Getwd()
	logger := log.New(os.Stderr, "", 0)

	os.MkdirAll(pwd+"/testing/Node-0", 0755)
	defer os.RemoveAll(pwd + "/testing")

	nodesBytes, _ := json.Marshal(config)
	ioutil.WriteFile(pwd+"/testing/Node-0/raftify.json", nodesBytes, 0755)

	node, err := InitNode(logger, pwd+"/testing/Node-0")
	if err != nil {
		t.Logf("Expected successful initialization of single-node cluster, instead got error: %v", err.Error())
		t.FailNow()
	}

	if node.GetState() != Leader {
		t.Logf("Expected node in single-node cluster to switch to leader immediately, instead it's in the %v state", node.state.toString())
		t.FailNow()
	}

	if err := node.Shutdown(); err != nil {
		t.Logf("Expected successful shutdown of Node_TestSingleNodeClusterWithNoPeers, instead got error: %v", err.Error())
		t.FailNow()
	}
}

func TestSingleNodeClusterWithPeers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestSingleNodeClusterWithPeers in short mode")
	}

	// Reserve ports for this test
	ports := reservePorts(2)

	config := Config{
		ID:       "Node_TestSingleNodeClusterWithPeers",
		MaxNodes: 2,
		Expect:   1,
		BindPort: ports[0],
		PeerList: []string{
			fmt.Sprintf("127.0.0.1:%v", ports[1]),
		},
	}

	// Initialize node.
	pwd, _ := os.Getwd()
	logger := log.New(os.Stderr, "", 0)

	os.MkdirAll(pwd+"/testing/Node-0", 0755)
	defer os.RemoveAll(pwd + "/testing")

	nodesBytes, _ := json.Marshal(config)
	ioutil.WriteFile(pwd+"/testing/Node-0/raftify.json", nodesBytes, 0755)

	node, err := InitNode(logger, pwd+"/testing/Node-0")
	if err != nil {
		t.Logf("Expected successful initialization of single-node cluster, instead got error: %v", err.Error())
		t.FailNow()
	}

	if node.GetState() == Leader {
		t.Log("Expected node in single-node cluster not to switch to leader immediately, instead it's in the leader state right away")
		t.FailNow()
	}

	if err := node.Shutdown(); err != nil {
		t.Logf("Expected successful shutdown of Node_TestSingleNodeClusterWithPeers, instead got error: %v", err.Error())
		t.FailNow()
	}
}

func TestNode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestNode in short mode")
	}

	// Reserve ports for this test
	ports := reservePorts(3)

	config := Config{
		ID:       "Node_TestNode",
		MaxNodes: 3,
		Expect:   3,
		BindPort: ports[0],
	}

	// Populate peerlist.
	for i := 0; i < config.MaxNodes; i++ {
		config.PeerList = append(config.PeerList, fmt.Sprintf("127.0.0.1:%v", ports[i]))
	}

	// Initialize all nodes.
	pwd, _ := os.Getwd()
	logger := log.New(os.Stderr, "", 0)
	nodes := []*Node{}

	for i := 0; i < config.MaxNodes; i++ {
		os.MkdirAll(fmt.Sprintf("%v/testing/Node-%v", pwd, i), 0755)
		defer os.RemoveAll(fmt.Sprintf("%v/testing", pwd))

		config.ID = fmt.Sprintf("Node-%v", i)
		config.BindPort = ports[i]

		nodesBytes, _ := json.Marshal(config)
		ioutil.WriteFile(fmt.Sprintf("%v/testing/Node-%v/raftify.json", pwd, i), nodesBytes, 0755)

		go func(pwd string, i int) {
			node, err := InitNode(logger, fmt.Sprintf("%v/testing/Node-%v", pwd, i))
			if err != nil {
				t.Logf("Expected successful initialization of Node-%v, instead got error: %v", i, err.Error())
				t.FailNow()
			}

			nodes = append(nodes, node)
		}(pwd, i)
	}

	// Wait for bootstrap to kick in for a leader to be elected.
	time.Sleep(2 * time.Second)

	// Check if every node is out of bootstrap mode.
	for i, node := range nodes {
		if node.state == Bootstrap {
			t.Logf("Expected Node-%v to be bootstrapped, instead it is still in bootstrap state", i)
			t.FailNow()
		}
	}

	// Test leader leave event.
	for i, node := range nodes {
		if node.state == Leader {
			if err := node.Shutdown(); err != nil {
				t.Logf("Expected successful shutdown of Node-%v, instead got error: %v", i, err.Error())
				t.FailNow()
			}
			nodes = append(nodes[:i], nodes[i+1:]...)
			break
		}
		if i == len(nodes)-1 {
			t.Log("Expected to find a leader and shut it down, instead couldn't find a leader")
			t.FailNow()
		}
	}

	// Wait for a new leader to be elected.
	time.Sleep(2 * time.Second)

	// Check if a new leader exists.
	for i, node := range nodes {
		if node.state == Leader {
			break
		}
		if i == len(nodes)-1 {
			t.Log("Expected a new leader to be elected after previous one left, instead couldn't find a leader")
			t.FailNow()
		}
	}

	for i, node := range nodes {
		if err := node.Shutdown(); err != nil {
			t.Logf("Expected successful shutdown of Node-%v, instead got error: %v", i, err.Error())
			t.FailNow()
		}
	}
}

func TestNodeRejoin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestNodeRejoin in short mode")
	}

	// Reserve ports for this test
	ports := reservePorts(3)

	config := Config{
		ID:       "Node_TestNodeRejoin",
		MaxNodes: 3,
		Expect:   2,
		BindPort: ports[0],
	}

	// Populate peerlist.
	for i := 0; i < config.MaxNodes; i++ {
		config.PeerList = append(config.PeerList, fmt.Sprintf("127.0.0.1:%v", ports[i]))
	}

	// Initialize all nodes except one normally.
	pwd, _ := os.Getwd()
	logger := log.New(os.Stderr, "", 0)
	nodes := []*Node{}

	for i := 0; i < config.MaxNodes; i++ {
		os.MkdirAll(fmt.Sprintf("%v/testing/Node-%v", pwd, i), 0755)
		defer os.RemoveAll(fmt.Sprintf("%v/testing", pwd))

		// Before the last node is initialized, the state.json file from another
		// node is first copied into its directory to trigger the rejoin.
		if i == config.MaxNodes-1 {
			// Wait for the other nodes to fully initialize.
			time.Sleep(2 * time.Second)

			src, err := os.Open(fmt.Sprintf("%v/testing/Node-%v/state.json", pwd, i-1))
			if err != nil {
				t.Logf("Expected state.json to be opened, instead got error: %v", err.Error())
				t.FailNow()
			}
			defer src.Close()

			dest, err := os.Create(fmt.Sprintf("%v/testing/Node-%v/state.json", pwd, i))
			if err != nil {
				t.Logf("Expected state.json to be created, instead got error: %v", err.Error())
				t.FailNow()
			}
			defer dest.Close()

			if _, err := io.Copy(dest, src); err != nil {
				t.Logf("Expected successful copy, instead got error: %v", err.Error())
				t.FailNow()
			}
		}

		config.ID = fmt.Sprintf("Node-%v", i)
		config.BindPort = ports[i]

		nodesBytes, _ := json.Marshal(config)
		ioutil.WriteFile(fmt.Sprintf("%v/testing/Node-%v/raftify.json", pwd, i), nodesBytes, 0755)

		go func(pwd string, i int) {
			node, err := InitNode(logger, fmt.Sprintf("%v/testing/Node-%v", pwd, i))
			if err != nil {
				t.Logf("Expected successful initialization of Node-%v, instead got error: %v", i, err.Error())
				t.FailNow()
			}

			nodes = append(nodes, node)
		}(pwd, i)
	}

	// Wait for bootstrap to kick in for a leader to be elected.
	time.Sleep(2 * time.Second)

	// Check whether the node has successfully rejoined.
	if nodes[config.MaxNodes-1].rejoin {
		t.Logf("Expected rejoin flag to be false, instead it's true")
		t.FailNow()
	}

	for _, node := range nodes {
		if err := node.Shutdown(); err != nil {
			t.Logf("Expected successful shutdown of Node_TestNode, instead got error: %v", err.Error())
			t.FailNow()
		}
	}
}
