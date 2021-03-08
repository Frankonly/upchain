package storage

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"upchain/crypto"
)

const (
	sizeConstantKey = "s"

	merklePrefix        = "m"
	leafHashIndexPrefix = "l"
	rootHashIndexPrefix = "r"
)

// HashPlaceHolder used to form a hash for calculating when there is no descendant fixed
const HashPlaceholder = "merkle placeholder"

type MerkleTreeStreaming struct {
	db    KvStore
	mutex sync.RWMutex // mutex protests not only database but also states in MerkleTreeStreaming

	// states
	root         InorderIndex
	rootHash     []byte
	lastHash     []byte
	next         uint64
	leftSiblings [maxLevel + 1][]byte
	isRootValid  bool

	// node placeholder
	placeholderHash []byte
}

// NewMerkleTreeStreaming is only used at beginning of upchain server.
// The db should be only used by one MerkleTreeStreaming, so there is no mutex used directly here.
func NewMerkleTreeStreaming(db KvStore) (MerkleAccumulator, error) {
	stream := &MerkleTreeStreaming{db: db}
	stream.placeholderHash = crypto.Hash([]byte(HashPlaceholder))

	res, err := db.Get(sizeKey())
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return nil, err
		}

		res = make([]byte, 8)
	}

	stream.next = binary.BigEndian.Uint64(res)
	if stream.next == 0 {
		if err := db.Put(sizeKey(), res); err != nil {
			return nil, err
		}

		stream.isRootValid = false
	} else {
		index := FromPostorder(stream.next - 1)

		hash, err := db.Get(merkleKey(index.Postorder()))
		if err != nil {
			return nil, err
		}

		// recover lost nodes
		for index.IsRightChild() {
			sibling := index.Sibling()
			siblingHash, err := db.Get(merkleKey(sibling.Postorder()))
			if err != nil {
				return nil, err
			}

			hash = crypto.HashNodes(siblingHash, hash)
			index = index.Parent()

			if err := db.Put(merkleKey(index.Postorder()), hash); err != nil {
				return nil, err
			}

			stream.next++
		}
		stream.lastHash = hash

		// update size
		if err := db.Put(sizeKeyValue(stream.next)); err != nil {
			return nil, err
		}

		// update left siblings
		lastLeaf := index.RightMostChild()
		rootLevel := RootLevelFromLeafIndex(lastLeaf.LeafIndexOnLevel())

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

		// update root
		_, err = stream.digest(false)
		if err != nil {
			return nil, err
		}
	}

	return stream, nil
}

// Get searches id in database layer to find its hash.
// Get only reads the database.
func (s *MerkleTreeStreaming) Get(id uint64) ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	index := FromLeafIndex(id)
	if index.Postorder() >= s.next {
		return nil, fmt.Errorf("%w: %d", ErrOutOfRange, id)
	}
	return s.db.Get(merkleKey(index.Postorder()))
}

// Append appends new hash to database layer.
// Append writes the database and states
func (s *MerkleTreeStreaming) Append(hash []byte) (uint64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	index := FromPostorder(s.next)
	if !index.IsLeaf() {
		return 0, fmt.Errorf("current position for writting is not a leaf")
	}

	s.isRootValid = false
	id := index.LeafIndexOnLevel()

	// using oldest proof strategy here
	_, err := s.db.Get(leafKey(hash))
	if errors.Is(err, ErrNotFound) {
		if err := s.db.Put(leafKeyValue(hash, index.Postorder())); err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	for i := range s.leftSiblings {
		if err := s.db.Put(merkleKey(index.Postorder()), hash); err != nil {
			return 0, err
		}
		s.next++

		if index.IsLeftChild() {
			s.leftSiblings[i] = hash
			s.lastHash = hash
			if s.root.Parent() == index || s.root == 0 {
				s.root = index
				s.rootHash = hash
				s.isRootValid = true
			}
			break
		}

		index = index.Parent()
		hash = crypto.HashNodes(s.leftSiblings[i], hash)
	}

	// update size
	if err := s.db.Put(sizeKeyValue(s.next)); err != nil {
		return 0, err
	}

	return id, nil
}

// Search searches hash in database layer to get the id of node. If there are several nodes contains the same hash,
// Search returns id of the oldest node (oldest strategy).
// Search reads and may write to database.
func (s *MerkleTreeStreaming) Search(hash []byte) (uint64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	value, err := s.db.Get(leafKey(hash))
	if err != nil {
		return 0, err
	}

	order := binary.BigEndian.Uint64(value)
	index := FromPostorder(order)
	if !index.IsLeaf() {
		return 0, fmt.Errorf("not leaf")
	}

	if index.Postorder() >= s.next {
		return 0, ErrNotFound
	}

	leafHash, err := s.db.Get(merkleKey(order))
	if err != nil {
		return 0, nil
	}

	if bytes.Compare(hash, leafHash) != 0 {
		err := s.db.Delete(leafKey(hash))
		if err != nil {
			return 0, err
		}

		return 0, ErrNotFound
	}

	return index.LeafIndexOnLevel(), nil
}

// Digest updates the root hash of Merkle tree and returns the root.
// Digest reads and may write to database and states.
func (s *MerkleTreeStreaming) Digest() ([]byte, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.digest(true)
}

// GetProof constructs a hash path who can proof the existence of data in certain id at the time of certain digest.
// GetProof reads and may write to database and states.
func (s *MerkleTreeStreaming) GetProof(id uint64, digest []byte) ([][]byte, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	index := FromLeafIndex(id)
	if index.Postorder() >= s.next {
		return nil, fmt.Errorf("%w: %d", ErrOutOfRange, id)
	}

	var err error
	var lastFrozen uint64
	var rootLevel int

	rootHash := digest
	if rootHash == nil {
		lastFrozen = s.next - 1

		// GetProof will return the latest digest, so the current root should be indexed
		rootHash, err = s.digest(true)
		if err != nil {
			return nil, err
		}

		rootLevel = s.root.Level()
	} else {
		value, err := s.db.Get(rootKey(rootHash))
		if errors.Is(err, ErrNotFound) {
			return nil, ErrInvalidDigest
		} else if err != nil {
			return nil, err
		}

		lastFrozen = binary.BigEndian.Uint64(value)
		lastIndex := FromPostorder(lastFrozen)
		rootLevel = RootLevelFromLeafIndex(lastIndex.RightMostChild().LeafIndexOnLevel())
	}

	if lastFrozen < index.Postorder() {
		return nil, ErrNotFound
	}

	hash, err := s.db.Get(merkleKey(index.Postorder()))
	if err != nil {
		return nil, err
	}

	if rootLevel == 0 {
		if bytes.Equal(rootHash, hash) {
			return [][]byte{rootHash}, nil
		}

		return nil, ErrNotFound
	}

	hashPath := make([][]byte, 0, rootLevel+2)
	hashPath = append(hashPath, hash)

	for index.Parent().Level() <= rootLevel {
		sibling := index.Sibling()
		siblingHash, err := s.getHash(sibling, lastFrozen)
		if err != nil {
			return nil, fmt.Errorf("failed to generate hash path: %s", err.Error())
		}

		if len(digest) == 0 {
			if index.IsLeftChild() {
				hash = crypto.HashNodes(hash, siblingHash)
			} else {
				hash = crypto.HashNodes(siblingHash, hash)
			}
		}

		hashPath = append(hashPath, siblingHash)
		index = index.Parent()
	}

	// check the validity of digest when using old digest
	if len(digest) == 0 && !bytes.Equal(rootHash, hash) {
		return nil, ErrInvalidDigest
	}

	hashPath = append(hashPath, rootHash)
	return hashPath, nil
}

func (s *MerkleTreeStreaming) Close() error {
	return s.db.Close()
}

// digest updates the root hash, reads and may write to database and states.
// mutex should be used when a function calls digest()
func (s *MerkleTreeStreaming) digest(indexRoot bool) ([]byte, error) {
	if !s.isRootValid {
		if s.next == 0 {
			return nil, ErrEmpty
		}

		index := FromPostorder(s.next - 1)
		hash := s.lastHash

		for index.LeftMostChild() != 0 {
			if index.IsLeftChild() {
				hash = crypto.HashNodes(hash, s.placeholderHash)
			} else {
				hash = crypto.HashNodes(s.leftSiblings[index.Level()], hash)
			}

			index = index.Parent()
		}

		s.root = index
		s.rootHash = hash
		s.isRootValid = true
	}

	if indexRoot {
		_, err := s.db.Get(rootKey(s.rootHash))
		if errors.Is(err, ErrNotFound) {
			order := FromPostorder(s.next - 1).Postorder()
			err = s.db.Put(rootKeyValue(s.rootHash, order))
		}

		if err != nil {
			return nil, err
		}
	}
	return s.rootHash, nil
}

// getHash reconstructs the node at the states with certain lastFrozen and returns the value.
func (s *MerkleTreeStreaming) getHash(index InorderIndex, lastFrozen uint64) ([]byte, error) {
	if index.Postorder() <= lastFrozen {
		return s.db.Get(merkleKey(index.Postorder()))
	}

	if index.LeftMostChild().Postorder() > lastFrozen {
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

	leftHash, err := s.getHash(leftChild, lastFrozen)
	if err != nil {
		return nil, err
	}

	rightHash, err := s.getHash(rightChild, lastFrozen)
	if err != nil {
		return nil, err
	}

	return crypto.HashNodes(leftHash, rightHash), nil
}
