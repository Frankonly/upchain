package storage

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"upchain/crypto"

	"github.com/stretchr/testify/require"
)

func TestMerkleTreeStreaming(t *testing.T) {
	r := require.New(t)

	path := filepath.Join(os.TempDir(), testDB)
	r.NoError(os.RemoveAll(path))

	db, err := NewLevelDB(path)
	r.NoError(err)
	r.NotNil(db)

	merkle, err := NewMerkleTreeStreaming(db)
	r.NoError(err)
	r.NotNil(merkle)

	_, err = merkle.Get(0)
	r.Error(err)
	_, err = merkle.Get(rand.Uint64())
	r.Error(err)

	r.NoError(merkle.Close())
	r.NoError(os.RemoveAll(path))
}

func TestMerkleTreeStreamingRW(t *testing.T) {
	r := require.New(t)
	rand.Seed(time.Now().UnixNano())

	path := filepath.Join(os.TempDir(), testDB)
	r.NoError(os.RemoveAll(path))

	db, err := NewLevelDB(path)
	r.NoError(err)
	r.NotNil(db)

	merkle, err := NewMerkleTreeStreaming(db)
	r.NoError(err)
	r.NotNil(merkle)

	hashes := make([][]byte, 1025)
	for i := range hashes {
		hashes[i] = make([]byte, 32)
		rand.Read(hashes[i])

		id, err := merkle.Append(hashes[i])
		r.NoError(err)
		r.EqualValues(i, id)

		hash, err := merkle.Get(id)
		r.NoError(err)
		r.Zero(bytes.Compare(hashes[i], hash))
	}

	for i := range hashes {
		value, err := merkle.Get(uint64(i))
		r.NoError(err)
		r.Equal(hashes[i], value)
	}

	r.NoError(merkle.Close())
	r.NoError(os.RemoveAll(path))
}

func TestMerkleTreeStreaming_Digest(t *testing.T) {
	r := require.New(t)
	rand.Seed(time.Now().UnixNano())

	path := filepath.Join(os.TempDir(), testDB)
	r.NoError(os.RemoveAll(path))

	db, err := NewLevelDB(path)
	r.NoError(err)
	r.NotNil(db)

	merkle, err := NewMerkleTreeStreaming(db)
	r.NoError(err)
	r.NotNil(merkle)

	hashes := make([][]byte, 1025)

	for i := range hashes {
		hashes[i] = make([]byte, 32)
		rand.Read(hashes[i])

		id, err := merkle.Append(hashes[i])
		r.NoError(err)
		r.EqualValues(i, id)

		value, err := merkle.Get(id)
		r.NoError(err)
		r.Zero(bytes.Compare(hashes[i], value))

		digest := testDigest(hashes[:i+1])
		r.Equal(digest, merkle.Digest())
	}

	r.NoError(merkle.Close())
	r.NoError(os.RemoveAll(path))
}

func TestMerkleTreeStreaming_GetProof(t *testing.T) {
	r := require.New(t)
	rand.Seed(time.Now().UnixNano())

	path := filepath.Join(os.TempDir(), testDB)
	r.NoError(os.RemoveAll(path))

	db, err := NewLevelDB(path)
	r.NoError(err)
	r.NotNil(db)

	merkle, err := NewMerkleTreeStreaming(db)
	r.NoError(err)
	r.NotNil(merkle)

	hashes := make([][]byte, 1025)

	for i := range hashes {
		hashes[i] = make([]byte, 32)
		rand.Read(hashes[i])

		id, err := merkle.Append(hashes[i])
		r.NoError(err)
		r.EqualValues(i, id)

		digest := merkle.Digest()
		r.NotEmpty(digest)

		id = rand.Uint64() % uint64(i+1)
		target, err := merkle.Get(id)
		r.NoError(err)

		path, err := merkle.GetProof(id)
		r.NoError(err)
		r.NotEmpty(path)

		r.Equal(target, path[0])
		r.Equal(digest, path[len(path)-1])
		r.True(testVerify(path[0], path[1:]))
	}

	r.NoError(merkle.Close())
	r.NoError(os.RemoveAll(path))
}

func TestMerkleTreeStreamingLoad(t *testing.T) {
	r := require.New(t)
	rand.Seed(time.Now().UnixNano())

	path := filepath.Join(os.TempDir(), testDB)
	r.NoError(os.RemoveAll(path))

	db, err := NewLevelDB(path)
	r.NoError(err)
	r.NotNil(db)

	merkle, err := NewMerkleTreeStreaming(db)
	r.NoError(err)
	r.NotNil(merkle)

	hashes := make([][]byte, 33)

	for i := range hashes {
		hashes[i] = make([]byte, 32)
		rand.Read(hashes[i])

		id, err := merkle.Append(hashes[i])
		r.NoError(err)
		r.EqualValues(i, id)

		digest := merkle.Digest()
		r.NotEmpty(digest)

		r.NoError(merkle.Close())

		db, err = NewLevelDB(path)
		r.NoError(err)
		r.NotNil(db)

		merkle, err = NewMerkleTreeStreaming(db)
		r.NoError(err)
		r.NotNil(merkle)

		target, err := merkle.Get(id)
		r.NoError(err)
		r.Equal(hashes[i], target)

		digest = merkle.Digest()
		r.NotEmpty(digest)

		path, err := merkle.GetProof(id)
		r.NoError(err)
		r.NotEmpty(path)

		r.Equal(target, path[0])
		r.Equal(digest, path[len(path)-1])
		r.True(testVerify(path[0], path[1:]))
	}

	r.NoError(merkle.Close())
	r.NoError(os.RemoveAll(path))
}

func TestMerkleTreeStreamingRecover(t *testing.T) {
	r := require.New(t)
	rand.Seed(time.Now().UnixNano())

	path := filepath.Join(os.TempDir(), testDB)
	r.NoError(os.RemoveAll(path))

	db, err := NewLevelDB(path)
	r.NoError(err)
	r.NotNil(db)

	merkle, err := NewMerkleTreeStreaming(db)
	r.NoError(err)
	r.NotNil(merkle)

	hashes := make([][]byte, 33)

	for i := range hashes {
		hashes[i] = make([]byte, 32)
		rand.Read(hashes[i])

		id, err := merkle.Append(hashes[i])
		r.NoError(err)
		r.EqualValues(i, id)

		digest := merkle.Digest()
		r.NotEmpty(digest)

		r.NoError(merkle.Close())

		db, err = NewLevelDB(path)
		r.NoError(err)
		r.NotNil(db)

		distance := FromLeafIndex(id+1).Postorder() - FromLeafIndex(id).Postorder()
		if distance > 1 {
			potentialSize := FromLeafIndex(id).Postorder() + 1 + uint64(rand.Intn(int(distance-1)))
			size := make([]byte, 8)
			binary.BigEndian.PutUint64(size, potentialSize)
			r.NoError(db.Put([]byte(sizeKey), size))
		}

		merkle, err = NewMerkleTreeStreaming(db)
		r.NoError(err)
		r.NotNil(merkle)

		target, err := merkle.Get(id)
		r.NoError(err)
		r.Equal(hashes[i], target)

		digest = merkle.Digest()
		r.NotEmpty(digest)

		path, err := merkle.GetProof(id)
		r.NoError(err)
		r.NotEmpty(path)

		r.Equal(target, path[0])
		r.Equal(digest, path[len(path)-1])
		r.True(testVerify(path[0], path[1:]))
	}

	r.NoError(merkle.Close())
	r.NoError(os.RemoveAll(path))
}

// testDigest works in a very slow way with O(n^2) complexity, only used for verifying the correctness
func testDigest(leaves [][]byte) []byte {
	if len(leaves) == 1 {
		return leaves[0]
	}

	parents := make([][]byte, len(leaves)/2+len(leaves)%2)
	for i := range parents {
		if 2*i+1 == len(leaves) {
			parents[i] = crypto.HashNodes(leaves[2*i], crypto.Hash([]byte(HashPlaceholder)))
		} else {
			parents[i] = crypto.HashNodes(leaves[2*i], leaves[2*i+1])
		}
	}

	return testDigest(parents)
}

// testVerify works in an extreme slow way with O(2^n) complexity, only used for verifying the correctness
func testVerify(target []byte, path [][]byte) bool {
	if len(path) == 0 {
		return true
	}

	if len(path) == 1 {
		return bytes.Equal(target, path[0])
	}

	if testVerify(crypto.HashNodes(target, path[0]), path[1:]) {
		return true
	}

	return testVerify(crypto.HashNodes(path[0], target), path[1:])
}
