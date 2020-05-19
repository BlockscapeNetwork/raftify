package raftify

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

// Helper function for writing raftify.json file.
func genConfig(node *Node) {
	jsonBytes, _ := json.Marshal(node.config)
	ioutil.WriteFile(node.workingDir+"/raftify.json", jsonBytes, 0755)
}

func TestConfigDefaults(t *testing.T) {
	pwd, _ := os.Getwd()
	node := &Node{
		logger:     new(log.Logger),
		workingDir: pwd,
		config: &Config{
			ID:       "Node_TestConfigDefaults",
			MaxNodes: 3,
			Expect:   1,
		},
	}

	genConfig(node)
	if err := node.loadConfig(); err != nil {
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

	os.Remove(pwd + "/raftify.json")
}

func TestLoadConfig(t *testing.T) {
	pwd, _ := os.Getwd()
	config := Config{
		ID:          "Node_TestLoadConfig",
		MaxNodes:    3,
		Encrypt:     "8ba4770b00f703fcc9e7d94f857db0e76fd53178d3d55c3e600a9f0fda9a75ad",
		Performance: 1,
		Expect:      1,
		LogLevel:    "DEBUG",
		BindAddr:    "0.0.0.0",
		BindPort:    3000,
		PeerList: []string{
			"0.0.0.0:3000",
			"0.0.0.0:3001",
			"0.0.0.0:3002",
		},
	}
	node := &Node{
		logger:     new(log.Logger),
		workingDir: pwd,
		config:     &config,
	}

	// Valid configuration.
	genConfig(node)
	if err := node.loadConfig(); err != nil {
		t.Logf("Expected valid configuration, instead got: %v", err.Error())
		t.Fail()
	}
	defer os.Remove(pwd + "/raftify.json")

	// Invalid ID.
	node.config.ID = ""
	genConfig(node)

	if err := node.loadConfig(); err == nil {
		t.Logf("Expected invalid ID, instead %v passed as valid", config.ID)
		t.Fail()
	}
	node.config.ID = "Node_TestLoadConfig"

	// Invalid maximum node limit.
	node.config.MaxNodes = 0
	genConfig(node)

	if err := node.loadConfig(); err == nil {
		t.Logf("Expected invalid maximum node limit, instead %v passed as valid", config.MaxNodes)
		t.Fail()
	}
	node.config.MaxNodes = 3

	// Invalid encryption key: Non-hex
	node.config.Encrypt = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	genConfig(node)

	if err := node.loadConfig(); err == nil {
		t.Logf("Expected invalid encryption key, instead %v passed as valid", config.Encrypt)
		t.Fail()
	}
	node.config.Encrypt = "8ba4770b00f703fcc9e7d94f857db0e76fd53178d3d55c3e600a9f0fda9a75ad"

	// Invalid performance.
	node.config.Performance = -1
	genConfig(node)

	if err := node.loadConfig(); err == nil {
		t.Logf("Expected invalid performance, instead %v passed as valid", config.Performance)
		t.Fail()
	}
	node.config.Performance = 1

	// Invalid number of expected nodes: Negative
	node.config.Expect = -1
	genConfig(node)

	if err := node.loadConfig(); err == nil {
		t.Logf("Expected invalid number of expected nodes, instead %v passed as valid", config.Expect)
		t.Fail()
	}
	node.config.Expect = 1

	// Invalid log level.
	node.config.LogLevel = "DEBUNK"
	genConfig(node)

	if err := node.loadConfig(); err == nil {
		t.Logf("Expected invalid log level, instead %v passed as valid", config.LogLevel)
		t.Fail()
	}
	node.config.LogLevel = "DEBUG"

	// Invalid IPv4.
	node.config.BindAddr = "192.168.500.213"
	genConfig(node)

	if err := node.loadConfig(); err == nil {
		t.Logf("Expected invalid bind address, instead %v passed as valid", config.BindAddr)
		t.Fail()
	}
	node.config.BindAddr = "0.0.0.0"

	// Invalid port: Too large
	node.config.BindPort = 123456
	genConfig(node)

	if err := node.loadConfig(); err == nil {
		t.Logf("Expected invalid bind port, instead %v passed as valid", config.BindPort)
		t.Fail()
	}

	// Invalid port: Too small
	node.config.BindPort = -1
	genConfig(node)

	if err := node.loadConfig(); err == nil {
		t.Logf("Expected invalid bind port, instead %v passed as valid", config.BindPort)
		t.Fail()
	}
	node.config.BindPort = 3000

	// Invalid peerlist: Wrong address format
	node.config.PeerList = []string{"192.168.500.213:3000", "192.168.0.213:123456", "192.168.0.213;3000"}
	genConfig(node)

	if err := node.loadConfig(); err == nil {
		t.Logf("Expected invalid peerlist, instead %v passed as valid", config.PeerList)
		t.Fail()
	}
	node.config.PeerList = []string{
		"0.0.0.0:3000",
		"0.0.0.0:3001",
		"0.0.0.0:3002",
	}

	// Invalid peerlist: Empty list and more than one node expected
	node.config.Expect = 2
	node.config.PeerList = []string{}
	genConfig(node)

	if err := node.loadConfig(); err == nil {
		t.Logf("Expected error for empty peerlist with more than one expected node, instead %v passed as valid", config.PeerList)
		t.Fail()
	}
	node.config.Expect = 1

	// Invalid peerlist: Too many peers for maximum nodes
	node.config.PeerList = []string{
		"0.0.0.0:3000", // Local node, will be truncated
		"0.0.0.0:3001",
		"0.0.0.0:3002",
		"0.0.0.0:3003",
	}

	if err := node.loadConfig(); err == nil {
		t.Logf("Expected invalid peerlist with too many peers, instead %v passed as valid", config.PeerList)
		t.Fail()
	}
	node.config.PeerList = []string{
		"0.0.0.0:3000",
		"0.0.0.0:3001",
		"0.0.0.0:3002",
	}
}

func TestTruncation(t *testing.T) {
	config := Config{
		ID:       "Node_TestTruncation",
		MaxNodes: 3,
		Expect:   1,
		BindAddr: "0.0.0.0",
		BindPort: 3000,
		PeerList: []string{
			"0.0.0.0:3000", // Local node
			"0.0.0.0:3001",
			"0.0.0.0:3002",
		},
	}

	if err := config.truncPeerList("0.0.0.0:3000"); err != nil {
		t.Logf("Expected truncation of peerlist, instead got: %v", err.Error())
		t.Fail()
	}
	if err := config.truncPeerList("0.0.0.0:3000"); err == nil {
		t.Logf("Expected no truncation of peerlist, instead %v:%v was truncated", config.BindAddr, config.BindPort)
		t.Fail()
	}
}
