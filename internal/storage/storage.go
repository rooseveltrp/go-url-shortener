package storage

import (
	"encoding/binary"
	"errors"

	"go.etcd.io/bbolt"
)

var (
	bucketURLs = []byte("urls")
	bucketHits = []byte("hits")

	ErrNotFound = errors.New("not found")
)

type Store struct {
	db *bbolt.DB
}

func New(path string) (*Store, error) {
	db, err := bbolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, err
	}
	if err := db.Update(func(tx *bbolt.Tx) error {
		if _, e := tx.CreateBucketIfNotExists(bucketURLs); e != nil {
			return e
		}
		if _, e := tx.CreateBucketIfNotExists(bucketHits); e != nil {
			return e
		}
		return nil
	}); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Save(code, url string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketURLs)
		return b.Put([]byte(code), []byte(url))
	})
}

func (s *Store) Exists(code string) (bool, error) {
	var ok bool
	err := s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket(bucketURLs).Get([]byte(code))
		ok = v != nil
		return nil
	})
	return ok, err
}

func (s *Store) Get(code string) (string, error) {
	var url []byte
	err := s.db.View(func(tx *bbolt.Tx) error {
		url = tx.Bucket(bucketURLs).Get([]byte(code))
		return nil
	})
	if err != nil {
		return "", err
	}
	if url == nil {
		return "", ErrNotFound
	}
	return string(url), nil
}

func (s *Store) IncHit(code string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketHits)
		key := []byte(code)
		raw := b.Get(key)
		var n uint64
		if raw != nil {
			n = binary.BigEndian.Uint64(raw)
		}
		n++
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, n)
		return b.Put(key, buf)
	})
}

func (s *Store) Hits(code string) (uint64, error) {
	var n uint64
	err := s.db.View(func(tx *bbolt.Tx) error {
		raw := tx.Bucket(bucketHits).Get([]byte(code))
		if raw != nil {
			n = binary.BigEndian.Uint64(raw)
		}
		return nil
	})
	return n, err
}
