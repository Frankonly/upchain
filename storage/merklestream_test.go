package storage

import (
	"bytes"
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

func TestMerkleTreeStreaming_RW(t *testing.T) {
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
