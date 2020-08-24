package raftify

import "math"

// toPreShutdown initiates the transition into a preshutdown node.
func (n *Node) toPreShutdown() {
	n.logger.Printf("[INFO] raftify: Preparing shutdown...")

	n.resetTimeout()
	n.messageTicker.Stop()

	n.state = PreShutdown
}

// runPreShutdown runs the preshutdown loop. This function is called within the runLoop function.
func (n *Node) runPreShutdown() {
	newQuorum := math.Ceil(float64(((len(n.memberlist.Members()) - 1) / 2) + 1))
	membersReached := n.sendNewQuorumToAll(int(newQuorum))

	// Make sure the new quorum can actually be reached after the node leaves
	if membersReached >= int(newQuorum) {
		n.toShutdown()
	}

	// Make sure a node in a single node cluster can leave appropriately
	if len(n.memberlist.Members()) == 1 {
		n.toShutdown()
	}

	// If not enough members could be reached runPreShutdown will just get called over and over
	// again in the runLoop function until the threshold is met for the node to switch into the
	// Shutdown state.
}
