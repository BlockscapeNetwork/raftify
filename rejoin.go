package raftify

// toRejoin initiates the transition into the rejoin state in case of a timeout or a
// crash-related node restart.
func (n *Node) toRejoin() {
	n.logger.Println("[INFO] raftify: Entering rejoin state")

	n.resetTimeout()
	n.messageTicker.Stop()

	n.state = Rejoin
}

// runRejoin runs the rejoin loop. This function is called within the runLoop function.
func (n *Node) runRejoin() {
	// Wait for the timeout timer to elapse
	<-n.timeoutTimer.C

	// Try rejoining the existing cluster via the peers in the peerlist
	if err := n.tryJoin(); err != nil {
		n.logger.Printf("[ERR] raftify: failed to rejoin cluster: %v\n", err.Error())
		n.resetTimeout()
		return
	}

	// On successful rejoin, switch into the Follower state
	n.logger.Printf("[INFO] raftify: %v successfully rejoined the cluster ✓", n.config.ID)
	n.toFollower(n.currentTerm)
}
