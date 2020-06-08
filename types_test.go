package raftify

import "testing"

func TestStateToString(t *testing.T) {
	state := Bootstrap
	if state.toString() != "Bootstrap" {
		t.Logf("Expected to get \"Bootstrap\", instead got %v", state.toString())
		t.Fail()
	}

	state = Follower
	if state.toString() != "Follower" {
		t.Logf("Expected to get \"Follower\", instead got %v", state.toString())
		t.Fail()
	}

	state = PreCandidate
	if state.toString() != "PreCandidate" {
		t.Logf("Expected to get \"PreCandidate\", instead got %v", state.toString())
		t.Fail()
	}

	state = Candidate
	if state.toString() != "Candidate" {
		t.Logf("Expected to get \"Candidate\", instead got %v", state.toString())
		t.Fail()
	}

	state = Leader
	if state.toString() != "Leader" {
		t.Logf("Expected to get \"Leader\", instead got %v", state.toString())
		t.Fail()
	}

	state = Shutdown
	if state.toString() != "Shutdown" {
		t.Logf("Expected to get \"Shutdown\", instead got %v", state.toString())
		t.Fail()
	}

	state = 100
	if state.toString() != "unknown" {
		t.Logf("Expected to get \"unknown\", instead got %v", state.toString())
		t.Fail()
	}
}

func TestMessageTypeToString(t *testing.T) {
	msg := HeartbeatMsg
	if msg.toString() != "HeartbeatMsg" {
		t.Logf("Expected to get \"HeartbeatMsg\", instead got %v", msg.toString())
		t.Fail()
	}

	msg = HeartbeatResponseMsg
	if msg.toString() != "HeartbeatResponseMsg" {
		t.Logf("Expected to get \"HeartbeatResponseMsg\", instead got %v", msg.toString())
		t.Fail()
	}

	msg = PreVoteRequestMsg
	if msg.toString() != "PreVoteRequestMsg" {
		t.Logf("Expected to get \"PreVoteRequestMsg\", instead got %v", msg.toString())
		t.Fail()
	}

	msg = PreVoteResponseMsg
	if msg.toString() != "PreVoteResponseMsg" {
		t.Logf("Expected to get \"PreVoteResponseMsg\", instead got %v", msg.toString())
		t.Fail()
	}

	msg = VoteRequestMsg
	if msg.toString() != "VoteRequestMsg" {
		t.Logf("Expected to get \"VoteRequestMsg\", instead got %v", msg.toString())
		t.Fail()
	}

	msg = VoteResponseMsg
	if msg.toString() != "VoteResponseMsg" {
		t.Logf("Expected to get \"VoteResponseMsg\", instead got %v", msg.toString())
		t.Fail()
	}

	msg = IntentionalLeaveMsg
	if msg.toString() != "IntentionalLeaveMsg" {
		t.Logf("Expected to get \"IntentionalLeaveMsg\", instead got %v", msg.toString())
		t.Fail()
	}

	msg = 100
	if msg.toString() != "unknown" {
		t.Logf("Expected to get \"unknown\", instead got %v", msg.toString())
		t.Fail()
	}
}
