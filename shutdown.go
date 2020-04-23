package raftify

import (
	"fmt"
)

// toShutdown initiates the transition into the shutdown mode. In this mode, the node
// leaves the cluster and shuts down gracefully while also removing the state.json file.
func (n *Node) toShutdown() {
	n.logger.Printf("[INFO] raftify: Shutting down %v...\n", n.config.ID)
	n.state = Shutdown
}

// runShutdown stops all timers/tickers and listenes, closes channels, leaves the memberlist
// and shuts down the node eventually.
func (n *Node) runShutdown() {
	n.timeoutTimer.Stop()
	n.messageTicker.Stop()

	n.deleteState()

	var errs string
	if err := n.memberlist.Leave(0); err != nil {
		errs += fmt.Sprintf("\t%v\n", err)
	}
	if err := n.memberlist.Shutdown(); err != nil {
		errs += fmt.Sprintf("\t%v\n", err)
	}

	if errs != "" {
		n.shutdownCh <- fmt.Errorf("found errors during shutdown:\n%v", errs)
		return
	}

	// Notify the shutdown channel so that the Shutdown API method can continue.
	n.shutdownCh <- nil
	n.logger.Println("[INFO] raftify: Shutdown successful ✓")
}
