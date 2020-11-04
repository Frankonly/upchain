package storage

import "github.com/syndtr/goleveldb/leveldb"

type LevelDBHelper struct {
	db *leveldb.DB
}

func NewLevelDB(name string) (KvStore, error) {
	db, err := leveldb.OpenFile(name, nil)
	if err != nil {
		return nil, err
	}

	return &LevelDBHelper{db: db}, nil
}

func (h *LevelDBHelper) Close() error {
	return h.db.Close()
}

func (h *LevelDBHelper) Get(key []byte) ([]byte, error) {
	return h.db.Get(key, nil)
}

func (h *LevelDBHelper) Put(key, value []byte) error {
	if err := h.db.Put(key, value, nil); err != nil {
		return err
	}

	return nil
}

func (h *LevelDBHelper) Delete(key []byte) error {
	if err := h.db.Delete(key, nil); err != nil {
		return err
	}

	return nil
}
