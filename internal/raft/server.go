package raft

import (
	"raft/internal/maths"
	"raft/internal/slices"
)

type State struct {
	currentTerm int
	votedFor    string
}

type Server struct {
	state       *State
	commitIndex int
	lastApplied int
	nextIndex   []int
	matchIndex  []int
	logStore    LogStore
}

func (server *Server) AppendEntries(
	leaderTerm int,
	leaderId int,
	prevLogIndex int,
	entries []Log,
	leaderCommit int,
) (term int, success bool) {
	if term < server.state.currentTerm {
		return server.state.currentTerm, false
	}

	log, err := server.logStore.Get(prevLogIndex)
	if err != nil {
		panic(err)
	}

	if log.term != leaderTerm {
		return server.state.currentTerm, false
	}

	logs, _ := server.logStore.GetAll()

	var i int
	entriesToAdd := make([]*Log, 0)
	for ; i < len(logs); i++ {
		for _, entry := range entries {
			if entry.IsConflictingWith(&logs[i]) {
				break
			}

			entriesToAdd = append(entriesToAdd, &entry)
			if entry.IsEqual(&logs[i]) {
				entriesToAdd = slices.RemoveByValue(entriesToAdd, &entry)
			}
		}
	}

	if i != len(logs) {
		server.logStore.DeleteFrom(i)
	}

	var maxIndex int
	for _, entry := range entriesToAdd {
		_, e := server.logStore.Append(entry)
		if e != nil {
			panic(e)
		}

		maxIndex = maths.Max(maxIndex, entry.index)
	}

	if leaderCommit > server.commitIndex {
		server.commitIndex = maths.Min(maxIndex, leaderCommit)
	}

	return server.state.currentTerm, true
}
