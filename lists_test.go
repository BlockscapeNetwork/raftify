package raftify

import (
	"testing"

	"github.com/hashicorp/memberlist"
)

func TestHeartbeatIDList(t *testing.T) {
	list := HeartbeatIDList{
		pending: []uint64{},
	}

	list.add(0)
	list.add(1)
	list.add(2)

	if len(list.pending) != 3 {
		t.Logf("Expected list length of 3, instead got %v", len(list.pending))
		t.Fail()
	}

	if err := list.remove(0); err != nil {
		t.Logf("Expected heartbeatid 0 to be removed, instead got error: %v", err.Error())
		t.Fail()
	}
	if len(list.pending) != 2 {
		t.Logf("Expected list length of 2, instead got %v", len(list.pending))
		t.Fail()
	}

	if err := list.remove(0); err == nil {
		t.Log("Expected removal of heartbeatid 0 to fail, instead passed as valid")
		t.Fail()
	}

	list.reset()
	if len(list.pending) != 0 {
		t.Logf("Expected list length of 0, instead got %v", len(list.pending))
		t.Fail()
	}
}

func TestVoteList(t *testing.T) {
	initPending := []*memberlist.Node{
		{Name: "1-One"},
		{Name: "2-Two"},
	}

	initPendingCopy := make([]*memberlist.Node, len(initPending))
	copy(initPendingCopy, initPending)

	list := VoteList{
		pending: initPendingCopy,
	}

	if err := list.remove("1-One"); err != nil {
		t.Logf("Expected to remove 1-One from vote list, instead got error: %v", err.Error())
		t.Fail()
	}
	if len(list.pending) != len(initPending)-1 {
		t.Logf("Expected list length of %v, instead got %v", len(initPending), len(list.pending))
		t.Fail()
	}

	if err := list.remove("2-Two"); err != nil {
		t.Logf("Expected to remove 2-Two from vote list, instead got error: %v", err.Error())
		t.Fail()
	}
	if len(list.pending) != len(initPending)-2 {
		t.Logf("Expected list length of %v, instead got %v", len(initPending), len(list.pending))
		t.Fail()
	}

	list.reset(initPending)
	if len(list.pending) != len(initPending) {
		t.Logf("Expected list to be reset to initial size, instead go size %v", len(initPending))
		t.Fail()
	}
}
