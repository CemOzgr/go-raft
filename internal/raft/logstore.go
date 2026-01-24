package raft

type LogStore interface {
	Append(log *Log) (int, error)
	Get(i int) (Log, error)
	GetAll() ([]Log, error)
	GetLastIndex() (int, error)
	DeleteFrom(i int)
}

type Log struct {
	Index   int
	Term    int
	Command []byte
}

func (log *Log) IsConflictingWith(other *Log) bool {
	return log.Index == other.Index && log.Term != other.Term
}

func (log *Log) IsEqual(other *Log) bool {
	return log.Index == other.Index && log.Term == other.Term
}
