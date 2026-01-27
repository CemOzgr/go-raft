package infra

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"raft/internal/convert"
	"raft/internal/raft"
	"slices"

	"github.com/cespare/xxhash"
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

	// length_index_term_commandLength_command_crc

	termBytes := convert.ToBytes(log.Term)
	indexBytes := convert.ToBytes(log.Index)
	commandBytes := []byte(log.Command)
	commandLengthBytes := convert.ToBytes(len(log.Command))

	payload := slices.Concat(
		indexBytes,
		termBytes,
		commandLengthBytes,
		commandBytes,
	)

	payload = binary.LittleEndian.AppendUint32(
		payload,
		uint32(len(payload)),
	)

	payload = binary.LittleEndian.AppendUint64(
		payload,
		xxhash.Sum64(payload),
	)

	f.Write(payload)

	return 0, nil
}

func (wal *WalLogStore) Get(i int) (raft.Log, error) {
	f, err := os.Open(path.Join(wal.logFolder, wal.fileName))
	if err != nil {
		return raft.Log{}, err
	}

	defer f.Close()

	var lengthBuffer [4]byte

	var current int
	for {
		f.Read(lengthBuffer[:])

		if current == i {
			break
		}

		f.Seek(
			int64(binary.LittleEndian.Uint64(lengthBuffer[:])),
			1,
		)

		current++
	}

	payloadBuffer := make(
		[]byte,
		binary.LittleEndian.Uint32(lengthBuffer[:]),
	)

	var crcBuffer [8]byte

	f.Read(payloadBuffer)
	f.Read(crcBuffer[:])

	return deserializeLog(payloadBuffer, crcBuffer[:])
}

func (wal *WalLogStore) GetAll() ([]raft.Log, error) {
	f, err := os.Open(path.Join(wal.logFolder, wal.fileName))
	if err != nil {
		return nil, err
	}

	defer f.Close()

	logs := make([]raft.Log, 0)

	var lengthBuffer [4]byte
	for {
		bytes, err := f.Read(lengthBuffer[:])
		if err != nil {
			return nil, err
		}

		if bytes == 0 {
			break
		}

		var crcBuffer [8]byte
		payloadBuffer := make(
			[]byte,
			binary.LittleEndian.Uint32(lengthBuffer[:]),
		)

		f.Read(payloadBuffer)
		f.Read(crcBuffer[:])

		log, err := deserializeLog(payloadBuffer, crcBuffer[:])
		if err != nil {
			return nil, err
		}

		logs = append(logs, log)
	}

	return logs, nil
}

func (wal *WalLogStore) GetLastIndex() (int, error) {
	f, err := os.Open(path.Join(wal.logFolder, wal.fileName))
	if err != nil {
		return 0, err
	}

	defer f.Close()

	var indexBuffer [4]byte
	var lengthBuffer [4]byte

	stat, _ := f.Stat()

	for {
		f.Read(lengthBuffer[:])
		f.Read(indexBuffer[:])
		if offset, _ := f.Seek(int64(binary.LittleEndian.Uint64(lengthBuffer[:]))-4+8, 1); offset >= stat.Size() {
			break
		}
	}

	return int(binary.LittleEndian.Uint32(indexBuffer[:])), nil
}

func (wal *WalLogStore) DeleteFrom(i int) {

}

func deserializeLog(buffer []byte, crcBuffer []byte) (raft.Log, error) {
	storedCrc := binary.LittleEndian.Uint64(crcBuffer[:])
	calculatedCrc := xxhash.Sum64(buffer)

	if storedCrc != calculatedCrc {
		return raft.Log{}, fmt.Errorf("crc mismatch: stored=%d, calculated=%d", storedCrc, calculatedCrc)
	}

	commandLength := binary.LittleEndian.Uint32(buffer[8:12])

	log := raft.Log{
		Index:   int(binary.LittleEndian.Uint32(buffer[:4])),
		Term:    int(binary.LittleEndian.Uint32(buffer[4:8])),
		Command: buffer[12 : 12+commandLength],
	}

	return log, nil
}
