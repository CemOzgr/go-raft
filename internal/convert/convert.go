package convert

import "encoding/binary"

func ToBytes(i int) []byte {
	bytes := make([]byte, 2)
	binary.LittleEndian.AppendUint16(bytes, uint16(i))

	return bytes
}
