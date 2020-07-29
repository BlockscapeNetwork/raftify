package raftify

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	pwd, _ := os.Getwd()
	node := &Node{
		logger:     new(log.Logger),
		workingDir: pwd,
		config: &Config{
			ID:       "TestNode",
			MaxNodes: 3,
			Expect:   1,
		},
	}

	genConfig(node)
	defer os.Remove(pwd + "/raftify.json")

	if err := node.loadConfig(false); err != nil {
		t.Logf("Expected valid configuration, instead got: %v", err.Error())
		t.Fail()
	}

	if node.config.Encrypt != "" {
		t.Logf("Expected encrypt to be empty, instead got %v", node.config.Encrypt)
		t.Fail()
	}
	if node.config.Performance != 1 {
		t.Logf("Expected performance to be 1, instead got %v", node.config.Performance)
		t.Fail()
	}
	if node.config.LogLevel != "WARN" {
		t.Logf("Expected log_level to be WARN, instead got %v", node.config.LogLevel)
		t.Fail()
	}
	if node.config.BindAddr != "0.0.0.0" {
		t.Logf("Expected bind_addr to bind to 0.0.0.0, instead got %v", node.config.BindAddr)
		t.Fail()
	}
	if node.config.BindPort != 7946 {
		t.Logf("Expected bind_port to bind to port 7946, instead got %v", node.config.BindPort)
		t.Fail()
	}
	if len(node.config.PeerList) != 0 {
		t.Logf("Expected peer_list to be empty, instead it has %v entries", len(node.config.PeerList))
		t.Fail()
	}
}

func TestLoadConfig(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(4)

	// Initialize dummy node
	node := initDummyNode("TestNode", 1, 3, ports[0])
	node.config.Encrypt = "8ba4770b00f703fcc9e7d94f857db0e76fd53178d3d55c3e600a9f0fda9a75ad"
	node.config.PeerList = []string{
		fmt.Sprintf("127.0.0.1:%v", ports[0]),
		fmt.Sprintf("127.0.0.1:%v", ports[1]),
		fmt.Sprintf("127.0.0.1:%v", ports[2]),
	}

	// Valid configuration.
	genConfig(node)
	defer os.Remove(node.workingDir + "/raftify.json")

	if err := node.loadConfig(false); err != nil {
		t.Logf("Expected valid configuration, instead got: %v", err.Error())
		t.Fail()
	}

	// Invalid ID.
	node.config.ID = ""
	genConfig(node)

	if err := node.loadConfig(false); err == nil {
		t.Logf("Expected invalid ID, instead %v passed as valid", node.config.ID)
		t.Fail()
	}
	node.config.ID = "TestNode"

	// Invalid maximum node limit.
	node.config.MaxNodes = 0
	genConfig(node)

	if err := node.loadConfig(false); err == nil {
		t.Logf("Expected invalid maximum node limit, instead %v passed as valid", node.config.MaxNodes)
		t.Fail()
	}
	node.config.MaxNodes = 3

	// Invalid encryption key: Non-hex
	node.config.Encrypt = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	genConfig(node)

	if err := node.loadConfig(false); err == nil {
		t.Logf("Expected invalid encryption key, instead %v passed as valid", node.config.Encrypt)
		t.Fail()
	}
	node.config.Encrypt = "8ba4770b00f703fcc9e7d94f857db0e76fd53178d3d55c3e600a9f0fda9a75ad"

	// Invalid performance.
	node.config.Performance = -1
	genConfig(node)

	if err := node.loadConfig(false); err == nil {
		t.Logf("Expected invalid performance, instead %v passed as valid", node.config.Performance)
		t.Fail()
	}
	node.config.Performance = 1

	// Invalid number of expected nodes: Negative
	node.config.Expect = -1
	genConfig(node)

	if err := node.loadConfig(false); err == nil {
		t.Logf("Expected invalid number of expected nodes, instead %v passed as valid", node.config.Expect)
		t.Fail()
	}
	node.config.Expect = 1

	// Invalid log level.
	node.config.LogLevel = "DEBUNK"
	genConfig(node)

	if err := node.loadConfig(false); err == nil {
		t.Logf("Expected invalid log level, instead %v passed as valid", node.config.LogLevel)
		t.Fail()
	}
	node.config.LogLevel = "DEBUG"

	// Invalid IPv4.
	node.config.BindAddr = "192.168.500.213"
	genConfig(node)

	if err := node.loadConfig(false); err == nil {
		t.Logf("Expected invalid bind address, instead %v passed as valid", node.config.BindAddr)
		t.Fail()
	}
	node.config.BindAddr = "127.0.0.1"

	// Invalid port: Too large
	node.config.BindPort = 123456
	genConfig(node)

	if err := node.loadConfig(false); err == nil {
		t.Logf("Expected invalid bind port, instead %v passed as valid", node.config.BindPort)
		t.Fail()
	}

	// Invalid port: Too small
	node.config.BindPort = -1
	genConfig(node)

	if err := node.loadConfig(false); err == nil {
		t.Logf("Expected invalid bind port, instead %v passed as valid", node.config.BindPort)
		t.Fail()
	}
	node.config.BindPort = ports[0]

	// Invalid peerlist: Wrong address format
	node.config.PeerList = []string{
		"192.168.500.213:6000",
		"192.168.0.213:123456",
		"192.168.0.213;7000",
	}
	genConfig(node)

	if err := node.loadConfig(false); err == nil {
		t.Logf("Expected invalid peerlist, instead %v passed as valid", node.config.PeerList)
		t.Fail()
	}
	node.config.PeerList = []string{
		fmt.Sprintf("127.0.0.1:%v", ports[0]),
		fmt.Sprintf("127.0.0.1:%v", ports[1]),
		fmt.Sprintf("127.0.0.1:%v", ports[2]),
	}

	// Invalid peerlist: Empty list and more than one node expected
	node.config.Expect = 2
	node.config.PeerList = []string{}
	genConfig(node)

	if err := node.loadConfig(false); err == nil {
		t.Logf("Expected error for empty peerlist with more than one expected node, instead %v passed as valid", node.config.PeerList)
		t.Fail()
	}
	node.config.Expect = 1

	// Invalid peerlist: Too many peers for maximum nodes
	node.config.PeerList = []string{
		fmt.Sprintf("127.0.0.1:%v", ports[0]), // Local node, will be truncated
		fmt.Sprintf("127.0.0.1:%v", ports[1]),
		fmt.Sprintf("127.0.0.1:%v", ports[2]),
		fmt.Sprintf("127.0.0.1:%v", ports[3]),
	}

	if err := node.loadConfig(false); err == nil {
		t.Logf("Expected invalid peerlist with too many peers, instead %v passed as valid", node.config.PeerList)
		t.Fail()
	}
	node.config.PeerList = []string{
		fmt.Sprintf("127.0.0.1:%v", ports[0]),
		fmt.Sprintf("127.0.0.1:%v", ports[1]),
		fmt.Sprintf("127.0.0.1:%v", ports[2]),
	}
}

func TestLoadConfigRejoin(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(3)

	// Initialize dummy nodes
	node1 := initDummyNode("TestNode_1", 2, 3, ports[0])
	node2 := initDummyNode("TestNode_2", 2, 3, ports[1])
	node3 := initDummyNode("TestNode_3", 1, 3, ports[2])

	genConfig(node1)
	genConfig(node2)
	genConfig(node3)
	defer os.Remove(node1.workingDir + "/raftify.json")
	defer os.Remove(node2.workingDir + "/raftify.json")
	defer os.Remove(node3.workingDir + "/raftify.json")

	node2.config.PeerList = []string{fmt.Sprintf("127.0.0.1:%v", node1.config.BindPort)}
	node3.config.PeerList = []string{fmt.Sprintf("127.0.0.1:%v", node1.config.BindPort)}

	// Starts nodes and form a cluster
	node1.createMemberlist()
	node2.createMemberlist()
	node3.createMemberlist()
	defer node1.memberlist.Shutdown()
	defer node2.memberlist.Shutdown()
	defer node3.memberlist.Shutdown()

	node2.tryJoin()
	node3.tryJoin()

	// Create dummy state.json file from node1's memberlist into the working directory
	node1.saveState()
	defer node1.deleteState()

	// Trigger rejoin and load the config
	node2.loadConfig(true)

	if len(node2.config.PeerList) != 2 {
		t.Logf("Expected peerlist of node2 to have two entries, instead got %v", len(node2.config.PeerList))
		t.FailNow()
	}
}

func TestTruncPeerList(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(3)

	// Initialize dummy node
	node := initDummyNode("TestNode", 1, 3, ports[0])
	node.config.PeerList = []string{
		fmt.Sprintf("127.0.0.1:%v", ports[0]),
		fmt.Sprintf("127.0.0.1:%v", ports[1]),
		fmt.Sprintf("127.0.0.1:%v", ports[2]),
	}

	if err := node.config.truncPeerList(fmt.Sprintf("127.0.0.1:%v", ports[0])); err != nil {
		t.Logf("Expected truncation of peerlist, instead got: %v", err.Error())
		t.Fail()
	}
	if err := node.config.truncPeerList(fmt.Sprintf("127.0.0.1:%v", ports[0])); err == nil {
		t.Logf("Expected no truncation of peerlist, instead %v:%v was truncated", node.config.BindAddr, node.config.BindPort)
		t.Fail()
	}
}
