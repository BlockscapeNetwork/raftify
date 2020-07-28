package raftify

// State is a custom type for all valid raftify node states.
type State uint8

// Constants for valid node states.
const (
	// Bootstrap is the state a node is in if it's waiting for the expected number of
	// nodes to go online before starting the Raft leader election.
	Bootstrap State = iota

	// Rejoin is the state a node is in if it times out or crashes and restarts.
	// In this state, it attempts to rejoin the existing cluster it dropped out of.
	Rejoin

	// Followers reset their timeout if they receive a heartbeat message from a leader.
	// If the timeout elapses, they become a precandidate.
	Follower

	// PreCandidates start collecting prevotes in order to determine if any other cluster
	// member has seen a leader and therefore make sure that there truly isn't one anymore
	// and a new one needs to be elected. Once the majority of prevotes have been granted,
	// it becomes a candidate.
	PreCandidate

	// Candidates enter a new election term and start collecting votes in order to be
	// promoted to the new cluster leader. A candidate votes for itself and waits for other
	// nodes to respond to its vote request. Sometimes a split vote can happen which means
	// that there are multiple candidates trying to become leader simultaneously such that
	// there are not enough votes left to reach quorum. In that case, the nodes wait for
	// the next timeout to start a new term. Once the majority of votes have been granted,
	// it becomes a leader.
	Candidate

	// The leader periodically sends out heartbeats to its followers to signal its availability.
	// If it suffers any sort of failure it automatically restarts as a follower. If it's
	// partitioned out and doesn't receive the majority of heartbeat responses it steps down.
	Leader

	// Shutdown is the state in which a node initiates a shutdown and gracefully allows the
	// runLoop goroutine to be exited and killed.
	Shutdown
)

// toString returns the string representation of a node state.
func (s *State) toString() string {
	switch *s {
	case Bootstrap:
		return "Bootstrap"
	case Rejoin:
		return "Rejoin"
	case Follower:
		return "Follower"
	case PreCandidate:
		return "PreCandidate"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	case Shutdown:
		return "Shutdown"
	default:
		return "unknown"
	}
}

// MessageType is a custom type for all valid raftify messages.
type MessageType uint8

// Constants for valid messages.
const (
	// A heartbeat message is sent out by a leader to signal availability.
	HeartbeatMsg MessageType = iota

	// A heartbeat response message is sent by the node who received the
	// heartbeat to the leader it originated from.
	HeartbeatResponseMsg

	// A prevote request message is sent out by aprecandidate in order to
	// make sure there truly is no more leader and a new candidacy needs to
	// be initiated.
	PreVoteRequestMsg

	// A prevote response message is sent by the node who received the
	// prevote request to the precandidate it originated from.
	PreVoteResponseMsg

	// A vote request message is sent out by a candidate in order to become
	// the new cluster leader.
	VoteRequestMsg

	// A vote response message is sent by the node who received the vote
	// request to the candidate it originated from.
	VoteResponseMsg
)

// toString returns the string representation of a message type.
func (t *MessageType) toString() string {
	switch *t {
	case HeartbeatMsg:
		return "HeartbeatMsg"
	case HeartbeatResponseMsg:
		return "HeartbeatResponseMsg"
	case PreVoteRequestMsg:
		return "PreVoteRequestMsg"
	case PreVoteResponseMsg:
		return "PreVoteResponseMsg"
	case VoteRequestMsg:
		return "VoteRequestMsg"
	case VoteResponseMsg:
		return "VoteResponseMsg"
	default:
		return "unknown"
	}
}
