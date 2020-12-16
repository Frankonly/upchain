package storage

import "fmt"

var (
	ErrOutOfRange    = fmt.Errorf("out of range")
	ErrNotFound      = fmt.Errorf("not found")
	ErrEmpty         = fmt.Errorf("empty")
	ErrInvalidDigest = fmt.Errorf("invliad digest")
)

type MerkleAccumulator interface {
	Append([]byte) (uint64, error)
	Get(uint64) ([]byte, error)
	Search([]byte) (uint64, error)
	Digest() ([]byte, error)
	GetProof(uint64, []byte) ([][]byte, error)
	Close() error
}

type KvStore interface {
	Get(key []byte) ([]byte, error)
	Put(key, value []byte) error
	Delete(key []byte) error
	Close() error
}
