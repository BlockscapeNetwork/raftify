package raftify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestMemberlist(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])

	if err := node.createMemberlist(); err != nil {
		t.Logf("Expected successful creation of memberlist, instead got error: %v", err.Error())
		t.FailNow()
	}
	if err := node.memberlist.Shutdown(); err != nil {
		t.Logf("Expected successful shutdown of memberlist, instead got error: %v", err.Error())
		t.FailNow()
	}
}

func TestTryJoin(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(2)

	// Initialize dummy nodes
	node1 := initDummyNode("TestNode_1", 1, 2, ports[0])
	node2 := initDummyNode("TestNode_2", 1, 2, ports[1])

	node1.config.PeerList = []string{fmt.Sprintf("127.0.0.1:%v", node2.config.BindPort)}
	node2.config.PeerList = []string{fmt.Sprintf("127.0.0.1:%v", node1.config.BindPort)}

	// Start node1 and fail while trying to join node2
	node1.createMemberlist()
	defer node1.memberlist.Shutdown()

	if err := node1.tryJoin(); err == nil {
		t.Logf("Expected node1 to throw an error on tryJoin, instead error was nil")
		t.FailNow()
	}

	// Start node2 and succeed while trying to join node1
	node2.createMemberlist()
	defer node2.memberlist.Shutdown()

	if err := node2.tryJoin(); err != nil {
		t.Logf("Expected node2 to successfully join node1, instead got error: %v", err.Error())
		t.FailNow()
	}
}

func ExampleNode_printMemberlist() {
	node := initDummyNode("TestNode", 1, 1, 4000)
	node.createMemberlist()
	node.printMemberlist()
	// Output:
	// [INFO] raftify: ->[] TestNode [127.0.0.1:4000] joined the cluster.
	// [INFO] raftify: The cluster has currently 1 members:
	// [INFO] raftify: - TestNode [127.0.0.1]
}

func TestInitNodeAndShutdown(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	tdir := fmt.Sprintf("%v/testing/TestNode", node.workingDir)

	os.MkdirAll(tdir, 0755)
	defer os.RemoveAll(fmt.Sprintf("%v/testing", node.workingDir))

	configBytes, _ := json.Marshal(node.config)
	ioutil.WriteFile(fmt.Sprintf("%v/raftify.json", tdir), configBytes, 0755)

	node, err := initNode(node.logger, tdir)
	if err != nil {
		t.Logf("Expected successful initialization of node, instead got error: %v", err.Error())
		t.FailNow()
	}
	if err = node.Shutdown(); err != nil {
		t.Logf("Expected successful shutdown of node, instead got error: %v", err.Error())
		t.FailNow()
	}
}

func TestResetTimeout(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.timeoutTimer = time.NewTimer(time.Second)
	node.timeoutTimer.Stop()

	start := time.Now()
	node.resetTimeout()
	end := <-node.timeoutTimer.C

	if end.Sub(start) > ((MaxTimeout*time.Duration(node.config.Performance)+10)*time.Millisecond) || time.Since(start) < (800*time.Duration(node.config.Performance)*time.Millisecond) {
		t.Logf("Expected timeout to elapse after %v-%vms, instead it took %vms", node.config.Performance*MinTimeout, node.config.Performance*MaxTimeout, time.Since(start))
		t.FailNow()
	}

	node.config.Performance = 2

	start = time.Now()
	node.resetTimeout()
	end = <-node.timeoutTimer.C

	if end.Sub(start) > ((MaxTimeout*time.Duration(node.config.Performance)+10)*time.Millisecond) || time.Since(start) < (800*time.Duration(node.config.Performance)*time.Millisecond) {
		t.Logf("Expected timeout to elapse after %v-%vms, instead it took %vms", node.config.Performance*MinTimeout, node.config.Performance*MaxTimeout, time.Since(start))
		t.FailNow()
	}
}

func TestStartMessageTicker(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.messageTicker = time.NewTicker(time.Millisecond)

	select {
	case <-node.messageTicker.C:
		node.messageTicker.Stop()
	case <-time.After(210 * time.Millisecond):
		t.Logf("Expected message ticker to have been called after 200ms, instead nothing happened")
		t.FailNow()
	}
}

func TestQuorum(t *testing.T) {
	// Reserve ports for this test
	ports := reservePorts(1)

	// Initialize dummy node
	node := initDummyNode("TestNode", 1, 1, ports[0])
	node.quorum = 2
	node.createMemberlist()

	if node.quorumReached(1) {
		t.Log("Expected first quorum check to return false, instead got true")
		t.FailNow()
	}
	if !node.quorumReached(2) {
		t.Log("Expected second quorum check to return true, instead got false")
		t.FailNow()
	}
	if node.quorum != 1 {
		t.Logf("Expected new quorum to be 1, instead got %v", node.quorum)
		t.FailNow()
	}

	if err := node.memberlist.Shutdown(); err != nil {
		t.Logf("Expected successful shutdown, instead got error: %v", err.Error())
		t.FailNow()
	}
}
