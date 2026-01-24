package raft

import (
	"math/rand"
	"raft/internal/maths"
	"raft/internal/slices"
)

type Volatile struct {
}

type State int

const (
	FOLLOWER State = iota
	CANDIDATE
	LEADER
)

type Server struct {
	state       State
	currentTerm int
	votedFor    int
	commitIndex int
	lastApplied int
	nextIndex   []int
	matchIndex  []int
	logStore    LogStore
	timeoutMs   int
}

func NewServer(
	logStore LogStore,
	term int,
	timeoutMin int,
	timeoutMax int,
) Server {
	server := Server{
		state:       FOLLOWER,
		currentTerm: term,
		votedFor:    -1,
		commitIndex: -1,
		lastApplied: -1,
		nextIndex:   make([]int, 5),
		matchIndex:  make([]int, 5),
		logStore:    logStore,
		timeoutMs:   rand.Intn(timeoutMax-timeoutMin+1) + timeoutMin,
	}

	return server
}

func (server *Server) AppendEntries(
	leaderTerm int,
	prevLogIndex int,
	entries []Log,
	leaderCommit int,
) (term int, success bool) {
	if term < server.currentTerm {
		return server.currentTerm, false
	}

	log, err := server.logStore.Get(prevLogIndex)
	if err != nil {
		panic(err)
	}

	if log.Term != leaderTerm {
		return server.currentTerm, false
	}

	logs, _ := server.logStore.GetAll()

	var i int
	for ; i < len(logs); i++ {
		for _, entry := range entries {
			if entry.IsConflictingWith(&logs[i]) {
				break
			}
		}
	}

	entriesToAdd := make([]*Log, 0)
	for ; i < len(logs); i++ {
		for _, entry := range entries {
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

		maxIndex = maths.Max(maxIndex, entry.Index)
	}

	if leaderCommit > server.commitIndex {
		server.commitIndex = maths.Min(maxIndex, leaderCommit)
	}

	return server.currentTerm, true
}

func (server *Server) RequestVote(
	term int,
	candidateId int,
	lastLogIndex int,
	lastLogTerm int,
) (currentTerm int, voteGranted bool) {
	if term < server.currentTerm {
		return server.currentTerm, false
	}

	if server.votedFor != -1 && server.votedFor != candidateId {
		return server.currentTerm, false
	}

	lastIndex, err := server.logStore.GetLastIndex()
	if err != nil {
		panic(err)
	}

	server.state = FOLLOWER
	return server.currentTerm, server.currentTerm <= lastLogTerm && lastIndex <= lastLogIndex
}
