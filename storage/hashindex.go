package storage

import "encoding/binary"

func leafKey(hash []byte) []byte {
	return append([]byte(leafHashIndexPrefix), hash...)
}

func leafKeyValue(hash []byte, order uint64) ([]byte, []byte) {
	value := make([]byte, 8)
	binary.BigEndian.PutUint64(value, order)

	return leafKey(hash), value
}
