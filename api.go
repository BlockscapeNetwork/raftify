package raftify

import "log"

// InitNode initializes a new raftified node.
func InitNode(logger *log.Logger, workingDir string) (*Node, error) {
	return initNode(logger, workingDir)
}

// Shutdown stops all timers/tickers and listeners, closes channels, leaves the
// memberlist and shuts down the node.
func (n *Node) Shutdown() error {
	defer close(n.messages.messageCh)
	defer close(n.events.eventCh)
	defer close(n.shutdownCh)

	n.shutdownCh <- nil
	err := <-n.shutdownCh
	return err
}

// GetHealthScore returns the health score according to memberlist. Lower numbers
// are better, and 0 means "totally healthy".
func (n *Node) GetHealthScore() int {
	return n.memberlist.GetHealthScore()
}

// GetMembers returns a map of the current memberlist with a key "id" and a
// value "address" in the host:port format.
func (n *Node) GetMembers() map[string]string {
	members := map[string]string{}
	for _, member := range n.memberlist.Members() {
		members[member.Name] = member.Address()
	}
	return members
}

// GetID returns the node's unique ID.
func (n *Node) GetID() string {
	return n.config.ID
}

// GetState returns the state the node's current state which is either Follower,
// PreCandidate, Candidate or Leader.
func (n *Node) GetState() State {
	return n.state
}
