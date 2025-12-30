package raft

type LogStore interface {
	Append(log *Log) (int, error)
	Get(i int) (Log, error)
	GetAll() ([]Log, error)
	GetLastIndex() (int, error)
	DeleteFrom(i int)
}

type Log struct {
	index   int
	term    int
	command []byte
}

func (log *Log) IsConflictingWith(other *Log) bool {
	return log.index == other.index && log.term != other.term
}

func (log *Log) IsEqual(other *Log) bool {
	return log.index == other.index && log.term == other.term
}
