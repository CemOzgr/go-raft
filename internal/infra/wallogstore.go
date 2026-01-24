package infra

import (
	"os"
	"path"
	"raft/internal/convert"
	"raft/internal/raft"
	"slices"
)

type WalLogStore struct {
	logFolder string
	signature []byte
	fileName  string
}

func (wal *WalLogStore) Append(log *raft.Log) (int, error) {
	f, err := os.OpenFile(
		path.Join(wal.logFolder, wal.fileName),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)

	if err != nil {
		return 0, err
	}

	defer f.Close()

	// term_index_commandLength_command

	termBytes := convert.ToBytes(log.Term)
	indexBytes := convert.ToBytes(log.Index)
	commandBytes := []byte(log.Command)
	commandLengthBytes := convert.ToBytes(len(log.Command))

	f.Write(slices.Concat(
		indexBytes,
		termBytes,
		commandLengthBytes,
		commandBytes,
	))

}

func (wal *WalLogStore) Get(i int) (raft.Log, error) {

}

func (wal *WalLogStore) GetAll() ([]raft.Log, error) {

}

func (wal *WalLogStore) GetLastIndex() (int, error) {

}

func (wal *WalLogStore) DeleteFrom(i int) {

}
