package storage

import (
	"errors"

	"github.com/syndtr/goleveldb/leveldb"
	lerrors "github.com/syndtr/goleveldb/leveldb/errors"
)

// LevelDBHelper is helper of leveldb.DB
type LevelDBHelper struct {
	db *leveldb.DB
}

// NewLevelDB news or opens a level DB from specified directory
func NewLevelDB(name string) (KvStore, error) {
	db, err := leveldb.OpenFile(name, nil)
	if err != nil {
		return nil, err
	}

	return &LevelDBHelper{db: db}, nil
}

// Close closes level DB
func (h *LevelDBHelper) Close() error {
	return h.db.Close()
}

// Get gets value from level DB
func (h *LevelDBHelper) Get(key []byte) ([]byte, error) {
	value, err := h.db.Get(key, nil)
	if errors.Is(err, lerrors.ErrNotFound) {
		return nil, ErrNotFound
	}

	return value, err
}

// Put puts a key-value to level DB
func (h *LevelDBHelper) Put(key, value []byte) error {
	if err := h.db.Put(key, value, nil); err != nil {
		return err
	}

	return nil
}

// Delete deletes a key-value from level DB
func (h *LevelDBHelper) Delete(key []byte) error {
	if err := h.db.Delete(key, nil); err != nil {
		return err
	}

	return nil
}
