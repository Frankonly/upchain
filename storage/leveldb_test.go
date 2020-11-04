package storage

import (
	"encoding/binary"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const testDB = "test.db"

func TestLevelDB(t *testing.T) {
	r := require.New(t)

	path := filepath.Join(os.TempDir(), testDB)
	r.NoError(os.RemoveAll(path))

	db, err := NewLevelDB(path)
	r.NoError(err)
	r.NotNil(db)

	r.NoError(db.(*LevelDBHelper).Close())
	r.NoError(os.RemoveAll(path))
}

func TestLevelDBRW(t *testing.T) {
	r := require.New(t)

	path := filepath.Join(os.TempDir(), testDB)
	r.NoError(os.RemoveAll(path))

	db, err := NewLevelDB(path)
	r.NoError(err)
	r.NotNil(db)

	key := []byte("test")
	value := []byte("Hello, LevelDB")

	_, err = db.Get(key)
	r.Error(err)
	r.NoError(db.Put(key, value))

	result, err := db.Get(key)
	r.NoError(err)
	r.Equal(value, result)

	r.NoError(db.Close())
	r.NoError(os.RemoveAll(path))
}

func TestLevelDBLoad(t *testing.T) {
	r := require.New(t)

	path := filepath.Join(os.TempDir(), testDB)
	r.NoError(os.RemoveAll(path))

	db, err := NewLevelDB(path)
	r.NoError(err)
	r.NotNil(db)

	kvMap := make(map[[8]byte][8]byte, 10000)

	for i := 0; i < 10000; i++ {
		key := [8]byte{}
		value := [8]byte{}
		binary.BigEndian.PutUint64(key[:], rand.Uint64())
		binary.BigEndian.PutUint64(value[:], rand.Uint64())

		kvMap[key] = value
		r.NoError(db.Put(key[:], value[:]))
	}

	r.NoError(db.Close())

	db, err = NewLevelDB(path)
	r.NoError(err)
	r.NotNil(db)

	for k, v := range kvMap {
		value, err := db.Get(k[:])
		r.NoError(err)
		r.Equal(v[:], value)
	}

	r.NoError(db.Close())
	r.NoError(os.RemoveAll(path))
}
