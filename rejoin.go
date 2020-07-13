package raftify

// toRejoin initiates the transition into the rejoin state in case of a timeout or a
// crash-related node restart.
func (n *Node) toRejoin(initialize bool) {
	n.logger.Println("[INFO] raftify: Entering rejoin state")

	n.resetTimeout()
	n.messageTicker.Stop()

	// If the initialize flag is set, it means that the rejoin is happening during node
	// initialization. This indicates that the runRejoin routine needs to send a signal
	// on the bootstrap channel in order to unblock InitNode.
	// If the flag is not set, it means that the rejoin is happening during operation
	// and it has already been bootstrapped once before. In that case, the runRejoin
	// routine is not going to signal anything on the bootstrap channel.
	if !initialize {
		n.state = Rejoin
	}
}

// runRejoin runs the rejoin loop. This function is called within the runLoop function.
func (n *Node) runRejoin() {
	// Wait for the timeout timer to elapse
	<-n.timeoutTimer.C

	// Try rejoining the existing cluster via the peers in the raftify.json
	if err := n.tryJoin(); err != nil {
		n.logger.Printf("[ERR] raftify: failed to rejoin cluster: %v\n", err.Error())
		n.resetTimeout()
		return
	}

	n.logger.Printf("[INFO] raftify: %v successfully rejoined the cluster ✓", n.config.ID)

	// Signal successful rejoin on bootstrap channel to unblock InitNode if the node
	// is currently initializing.
	if n.state == Initialize {
		n.bootstrapCh <- true
	}

	// On successful rejoin, switch into the Follower state
	n.toFollower(n.currentTerm)
}
