package upchain

import (
	"encoding/binary"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
)

type MerkleAccumulator interface {
	Append([]byte) (uint64, error)
	Get(uint64) ([]byte, error)
	Digest() []byte
	GetProof(uint64) ([][]byte, error)
}

type KvStore interface {
	Get(key []byte) ([]byte, error)
	Put(key, value []byte) error
	Delete(key []byte) error
}

type LevelDBHelper struct {
	db *leveldb.DB
}

func NewLevelDB(name string) (KvStore, error) {
	db, err := leveldb.OpenFile(name, nil)
	if err != nil {
		return nil, err
	}

	return LevelDBHelper{db: db}, nil
}

func (h LevelDBHelper) Close() error {
	return h.db.Close()
}

func (h LevelDBHelper) Get(key []byte) ([]byte, error) {
	return h.db.Get(key, nil)
}

func (h LevelDBHelper) Put(key, value []byte) error {
	if err := h.db.Put(key, value, nil); err != nil {
		return err
	}

	return nil
}

func (h LevelDBHelper) Delete(key []byte) error {
	if err := h.db.Delete(key, nil); err != nil {
		return err
	}

	return nil
}

const (
	sizeKey      = "s"
	merklePrefix = "m"
)

const PlaceholderHash = "merkle placeholder"

type MerkleTreeStreaming struct {
	db              KvStore
	root            InorderIndex
	rootHash        []byte
	lastHash        []byte
	next            uint64
	leftSiblings    [maxLevel + 1][]byte
	isRootValid     bool
	placeholderHash []byte
}

func NewMerkleTreeStreaming(db KvStore) (MerkleAccumulator, error) {
	stream := &MerkleTreeStreaming{db: db}
	stream.placeholderHash = Hash([]byte(PlaceholderHash))

	res, err := db.Get([]byte(sizeKey))
	if err != nil {
		return nil, err
	}

	stream.next = binary.BigEndian.Uint64(res)
	if stream.next == 0 {
		stream.isRootValid = false
	} else {
		index := FromPostorder(stream.next - 1)

		hash, err := db.Get(merkleKey(index.Postorder()))
		if err != nil {
			return nil, err
		}

		for index.IsRightChild() {
			sibling := index.Sibling()
			siblingHash, err := db.Get(merkleKey(sibling.Postorder()))
			if err != nil {
				return nil, err
			}

			hash = HashNodes(siblingHash, hash)
			index = index.Parent()

			if err := db.Put(merkleKey(index.Postorder()), hash); err != nil {
				return nil, err
			}

			stream.next++
		}
		stream.lastHash = hash

		// update size
		size := make([]byte, 8)
		binary.BigEndian.PutUint64(size, stream.next)
		if err := db.Put([]byte(sizeKey), size); err != nil {
			return nil, err
		}

		lastLeaf := index.RightMostChild()
		rootLevel := RootLevelFromLeaves(lastLeaf.LeavesOnLevel())

		for index.Level() <= rootLevel {
			// judge whether the node is frozen
			if index.Postorder() < stream.next {
				// frozen node here must be left child
				hash, err := db.Get(merkleKey(index.Postorder()))
				if err != nil {
					return nil, err
				}

				stream.leftSiblings[index.Level()] = hash
			} else {
				if index.IsRightChild() {
					// left sibling here must be frozen node
					index = index.Sibling()
					hash, err := db.Get(merkleKey(index.Postorder()))
					if err != nil {
						return nil, err
					}

					stream.leftSiblings[index.Level()] = hash
				}
			}

			index = index.Parent()
		}

		root, err := db.Get(merkleKey(index.Postorder()))
		if err != nil {
			return nil, err
		}

		stream.root = index
		stream.rootHash = root
	}
	return stream, nil
}

func (s MerkleTreeStreaming) Get(id uint64) ([]byte, error) {
	index := FromLeaves(id)
	if index.Postorder() >= s.next {
		return nil, fmt.Errorf("id out of range: %d", id)
	}
	return s.db.Get(merkleKey(index.Postorder()))
}

func (s MerkleTreeStreaming) Append(hash []byte) (uint64, error) {
	index := FromPostorder(s.next)
	if !index.IsLeaf() {
		return 0, fmt.Errorf("current position for writting is not a leaf")
	}

	s.isRootValid = false
	id := index.LeavesOnLevel()

	for i := range s.leftSiblings {
		if err := s.db.Put(merkleKey(index.Postorder()), hash); err != nil {
			return 0, err
		}
		s.next++

		if index.IsLeftChild() {
			s.leftSiblings[i] = hash
			s.lastHash = hash
			if s.root.Parent() == index {
				s.root = index
				s.rootHash = hash
				s.isRootValid = true
			}
			break
		}

		index = index.Parent()
		hash = HashNodes(s.leftSiblings[i], hash)
	}

	// update size
	size := make([]byte, 8)
	binary.BigEndian.PutUint64(size, s.next)
	if err := s.db.Put([]byte(sizeKey), size); err != nil {
		return 0, err
	}

	return id, nil
}

func (s MerkleTreeStreaming) Digest() []byte {
	if !s.isRootValid {
		index := FromPostorder(s.next - 1)
		hash := s.lastHash
		rootLevel := s.root.Level()

		for index.Level() < rootLevel {
			if index.IsLeftChild() {
				hash = HashNodes(hash, s.placeholderHash)
			} else {
				hash = HashNodes(s.leftSiblings[index.Level()], hash)
			}

			index = index.Parent()
		}

		s.rootHash = hash
	}

	return s.rootHash
}

func (s MerkleTreeStreaming) GetProof(id uint64) ([][]byte, error) {
	index := FromLeaves(id)

	if index.Postorder() >= s.next {
		return nil, fmt.Errorf("id out of range: %d", id)
	}

	hash, err := s.db.Get(merkleKey(index.Postorder()))
	if err != nil {
		return nil, err
	}

	rootLevel := s.root.Level()
	hashPath := make([][]byte, 0, rootLevel+1)
	hashPath = append(hashPath, hash)

	for index.Parent().Level() <= rootLevel {
		sibling := index.Sibling()
		siblingHash, err := s.getCurrentHash(sibling)
		if err != nil {
			return nil, err
		}

		hashPath = append(hashPath, siblingHash)
	}

	hashPath = append(hashPath, s.Digest())
	return hashPath, nil
}

func (s MerkleTreeStreaming) getCurrentHash(index InorderIndex) ([]byte, error) {
	if index.Postorder() < s.next {
		return s.db.Get(merkleKey(index.Postorder()))
	}

	if index.LeftMostChild().Postorder() >= s.next {
		return s.placeholderHash, nil
	}

	leftChild, err := index.LeftChild()
	if err != nil {
		return nil, err
	}

	rightChild, err := index.RightChild()
	if err != nil {
		return nil, err
	}

	leftHash, err := s.getCurrentHash(leftChild)
	if err != nil {
		return nil, err
	}

	rightHash, err := s.getCurrentHash(rightChild)
	if err != nil {
		return nil, err
	}

	return HashNodes(leftHash, rightHash), nil
}

func merkleKey(id uint64) []byte {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, id)
	return append([]byte(merklePrefix), key...)
}
