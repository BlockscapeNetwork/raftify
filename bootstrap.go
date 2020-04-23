package raftify

import (
	"time"
)

// tryJoin attempts to join an existing cluster via one of its peers listed in the peerlist.
// If no peers can be reached the node is started and waits to be bootstrapped.
func (n *Node) tryJoin() error {
	n.logger.Println("[DEBUG] raftify: Trying to join existing cluster via peers...")
	numPeers, err := n.memberlist.Join(n.config.PeerList)
	if err != nil {
		return err
	}

	n.logger.Printf("[DEBUG] raftify: %v peers are currently available ✓\n", numPeers)
	return nil
}

// toBootstrap initiates the transition into the bootstrap mode. In this mode, nodes wait for
// the expected number of nodes specified in the expect field of the raftify.json to go online
// and start all nodes of the cluster at the same time.
func (n *Node) toBootstrap() {
	n.logger.Printf("[DEBUG] raftify: %v/%v nodes for bootstrap...\n", len(n.memberlist.Members()), n.config.Expect)
	n.state = Bootstrap

	if n.config.Expect == 1 {
		n.logger.Println("[DEBUG] raftify: Successfully bootstrapped cluster ✓")
		n.saveState()
		n.toLeader()
		n.printMemberlist()
		return
	}

	if err := n.tryJoin(); err != nil {
		n.logger.Printf("[ERR] raftify: failed to join cluster: %v\nTrying again...\n", err.Error())
	}
}

// runBootstrap waits for the number of nodes specified in the expect field of the raftify.json
// to join the cluster. This function is called within the runLoop function.
func (n *Node) runBootstrap() {
	select {
	case <-n.events.eventCh:
		n.logger.Printf("[DEBUG] raftify: %v/%v nodes for bootstrap...\n", len(n.memberlist.Members()), n.config.Expect)
		n.printMemberlist()
		n.saveState()

		if len(n.memberlist.Members()) >= n.config.Expect {
			n.logger.Println("[DEBUG] raftify: Successfully bootstrapped cluster ✓")
			n.toFollower(0)
		}

	case <-time.After(5 * time.Second):
		if err := n.tryJoin(); err != nil {
			n.logger.Printf("[ERR] raftify: failed to join cluster: %v\nTrying again...\n", err.Error())
		}

	case <-n.shutdownCh:
		n.toShutdown()
	}
}
