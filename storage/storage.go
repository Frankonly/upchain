package storage

import "fmt"

var (
	// ErrOutOfRange indicates that request is out of range
	ErrOutOfRange = fmt.Errorf("out of range")
	// ErrNotFound indicates that key is not found
	ErrNotFound = fmt.Errorf("not found")
	// ErrEmpty indicates that there is no data
	ErrEmpty = fmt.Errorf("empty")
	// ErrInvalidDigest indicates that digest is invalid
	ErrInvalidDigest = fmt.Errorf("invliad digest")
)

// MerkleAccumulator defines core operations of merkle accumulator
type MerkleAccumulator interface {
	Append([]byte) (uint64, error)
	Get(uint64) ([]byte, error)
	Search([]byte) (uint64, error)
	Digest() ([]byte, error)
	GetProof(uint64, []byte) ([][]byte, error)
	Close() error
}

// KvStore supports basic functions of kv store
type KvStore interface {
	Get(key []byte) ([]byte, error)
	Put(key, value []byte) error
	Delete(key []byte) error
	Close() error
}
