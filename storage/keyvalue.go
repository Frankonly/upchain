package storage

import (
	"encoding/binary"
)

func merkleKey(order uint64) []byte {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, order)

	return append([]byte(merklePrefix), key...)
}

func sizeKey() []byte {
	return []byte(sizeConstantKey)
}

func sizeKeyValue(size uint64) ([]byte, []byte) {
	sizeValue := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeValue, size)

	return []byte(sizeConstantKey), sizeValue
}

func leafKey(hash []byte) []byte {
	return append([]byte(leafHashIndexPrefix), hash...)
}

func leafKeyValue(hash []byte, order uint64) ([]byte, []byte) {
	value := make([]byte, 8)
	binary.BigEndian.PutUint64(value, order)

	return leafKey(hash), value
}
