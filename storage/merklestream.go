package storage

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"upchain/crypto"
)

const (
	sizeConstantKey = "s"

	merklePrefix        = "m"
	leafHashIndexPrefix = "l"
	rootHashIndexPrefix = "r"
)

const HashPlaceholder = "merkle placeholder"

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

func (s *MerkleTreeStreaming) Get(id uint64) ([]byte, error) {
	index := FromLeafIndex(id)
	if index.Postorder() >= s.next {
		return nil, fmt.Errorf("%w: %d", ErrOutOfRange, id)
	}
	return s.db.Get(merkleKey(index.Postorder()))
}

func (s *MerkleTreeStreaming) Append(hash []byte) (uint64, error) {
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

func (s *MerkleTreeStreaming) Search(hash []byte) (uint64, error) {
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

func (s *MerkleTreeStreaming) Digest() ([]byte, error) {
	return s.digest(true)
}

func (s *MerkleTreeStreaming) GetProof(id uint64) ([][]byte, error) {
	index := FromLeafIndex(id)

	if index.Postorder() >= s.next {
		return nil, fmt.Errorf("%w: %d", ErrOutOfRange, id)
	}

	rootHash, err := s.digest(false)
	if err != nil {
		return nil, err
	}

	rootLevel := s.root.Level()
	if rootLevel == 0 {
		return [][]byte{rootHash}, nil
	}

	hash, err := s.db.Get(merkleKey(index.Postorder()))
	if err != nil {
		return nil, err
	}

	hashPath := make([][]byte, 0, rootLevel+2)
	hashPath = append(hashPath, hash)

	for index.Parent().Level() <= rootLevel {
		sibling := index.Sibling()
		siblingHash, err := s.getCurrentHash(sibling)
		if err != nil {
			return nil, fmt.Errorf("failed to generate hash path: %s", err.Error())
		}

		hashPath = append(hashPath, siblingHash)
		index = index.Parent()
	}

	hashPath = append(hashPath, rootHash)
	return hashPath, nil
}

func (s *MerkleTreeStreaming) Close() error {
	return s.db.Close()
}

func (s *MerkleTreeStreaming) digest(indexRoot bool) ([]byte, error) {
	if !s.isRootValid {
		if s.next == 0 {
			return nil, ErrEmpty
		}

		index := FromPostorder(s.next - 1)
		hash := s.lastHash
		order := index.RightMostChild().Postorder()

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

		if indexRoot {
			if err := s.db.Put(rootKeyValue(hash, order)); err != nil {
				return nil, err
			}
		}
	}

	return s.rootHash, nil
}

func (s *MerkleTreeStreaming) getCurrentHash(index InorderIndex) ([]byte, error) {
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

	return crypto.HashNodes(leftHash, rightHash), nil
}
