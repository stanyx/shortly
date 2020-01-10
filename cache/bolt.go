package cache

import (
	"errors"
	"log"

	bolt "go.etcd.io/bbolt"
)

const cacheBucketName = "urls"

type BoltDBCache struct {
	db     *bolt.DB
	logger *log.Logger
}

func NewBoltDBCache(db *bolt.DB, logger *log.Logger) (*BoltDBCache, error) {

	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(cacheBucketName))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &BoltDBCache{
		db:     db,
		logger: logger,
	}, nil
}

func (ch *BoltDBCache) Load(key interface{}) (interface{}, bool) {

	var value []byte

	err := ch.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(cacheBucketName))
		value = b.Get([]byte(key.(string)))
		if len(value) == 0 {
			return errors.New("key miss")
		}
		return nil
	})
	if err != nil {
		ch.logger.Printf("cache get(key=%v) error, cause: %+v\n", key, err)
		return nil, false
	}
	return string(value), true
}

func (ch *BoltDBCache) Store(key interface{}, value interface{}) {

	err := ch.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(cacheBucketName))
		err := b.Put([]byte(key.(string)), []byte(value.(string)))
		return err
	})

	if err != nil {
		ch.logger.Printf("cache set(key=%v) error, cause: %+v\n", key, err)
	}
}

func (ch *BoltDBCache) Delete(key interface{}) {
	err := ch.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(cacheBucketName))
		err := b.Delete([]byte(key.(string)))
		return err
	})
	if err != nil {
		ch.logger.Printf("cache delete(key=%s) error, cause: %+v\n", key, err)
	}
}

func (ch *BoltDBCache) Range(f func(key interface{}, value interface{}) bool) {
	err := ch.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(cacheBucketName))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			f(k, v)
		}
		return nil
	})
	if err != nil {
		ch.logger.Printf("cache range error, cause: %+v\n", err)
	}
}

func (ch *BoltDBCache) Ping() error {
	return nil
}
