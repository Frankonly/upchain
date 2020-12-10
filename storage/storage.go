package storage

import "fmt"

var (
	ErrOutOfRange = fmt.Errorf("out of range")
	ErrNotFound   = fmt.Errorf("not found")
)

type MerkleAccumulator interface {
	Append([]byte) (uint64, error)
	Get(uint64) ([]byte, error)
	Search([]byte) (uint64, error)
	Digest() []byte
	GetProof(uint64) ([][]byte, error)
	Close() error
}

type KvStore interface {
	Get(key []byte) ([]byte, error)
	Put(key, value []byte) error
	Delete(key []byte) error
	Close() error
}
