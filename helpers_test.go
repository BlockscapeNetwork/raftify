package raftify

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/memberlist"
)

var (
	mutex      = sync.Mutex{}
	unusedPort = 3000
)

// reservePorts returns a slice of unused ports.
func reservePorts(number int) []int {
	mutex.Lock()
	defer mutex.Unlock()

	var ports []int
	for i := unusedPort; i < (unusedPort + number); i++ {
		ports = append(ports, i)
	}

	unusedPort += number
	return ports
}

func TestReservePorts(t *testing.T) {
	ports1 := reservePorts(10)
	if len(ports1) != 10 {
		t.Logf("Expected 10 ports to have been reserved, instead got %v", len(ports1))
		t.FailNow()
	}

	ports2 := reservePorts(10)
	if len(ports2) != 10 {
		t.Logf("Expected 10 ports to have been reserved, instead got %v", len(ports2))
		t.FailNow()
	}

	if ports1[9]+1 != ports2[0] {
		t.Logf("Expected the second set of ports to immediately follow up the first one")
		t.FailNow()
	}
}

// Helper function for writing raftify.json file.
func genConfig(node *Node) {
	jsonBytes, _ := json.Marshal(node.config)
	ioutil.WriteFile(node.workingDir+"/raftify.json", jsonBytes, 0755)
}

// initDummyNode initializes a node.
func initDummyNode(id string, expect, maxnodes, port int) *Node {
	logger := log.New(os.Stdout, "", 0)
	pwd, _ := os.Getwd()
	node := &Node{
		logger:     logger,
		workingDir: pwd,
		config: &Config{
			ID:          id,
			MaxNodes:    maxnodes,
			Expect:      expect,
			Performance: 1,
			LogLevel:    "DEBUG",
			BindAddr:    "127.0.0.1",
			BindPort:    port,
		},
		messages: &MessageDelegate{
			logger:    logger,
			messageCh: make(chan []byte),
		},
		events: &ChannelEventDelegate{
			logger:  logger,
			eventCh: make(chan memberlist.NodeEvent, maxnodes),
		},
		timeoutTimer:  time.NewTimer(time.Second),
		messageTicker: time.NewTicker(time.Second),
		bootstrapCh:   make(chan bool),
		shutdownCh:    make(chan error),
		heartbeatIDList: &HeartbeatIDList{
			logger:             logger,
			currentHeartbeatID: 0,
			received:           0,
			pending:            []uint64{},
			subQuorumCycles:    0,
		},
		preVoteList: &VoteList{
			logger:              logger,
			received:            0,
			pending:             []*memberlist.Node{},
			missedPrevoteCycles: 0,
		},
		voteList: &VoteList{
			logger:              logger,
			received:            0,
			pending:             []*memberlist.Node{},
			missedPrevoteCycles: 0,
		},
	}

	node.timeoutTimer.Stop()
	node.messageTicker.Stop()
	return node
}
