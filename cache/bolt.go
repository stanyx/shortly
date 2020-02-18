package cache

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

const cacheBucketName = "urls"

type BoltDBCache struct {
	sync.WaitGroup
	db     *bolt.DB
	logger *log.Logger
	stopCh chan struct{}
}

func NewBoltDBCache(db *bolt.DB, logger *log.Logger) (*BoltDBCache, error) {

	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(cacheBucketName))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(cacheBucketName + ":ttl"))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	stopCh := make(chan struct{}, 1)
	ch := &BoltDBCache{
		db:     db,
		logger: logger,
		stopCh: stopCh,
	}
	ch.Add(1)
	// expired records clean loop
	bucketData := []byte(cacheBucketName)
	bucketTTL := []byte(cacheBucketName + ":ttl")
	go func() {
	loop:
		for {
			select {
			case <-stopCh:
				break loop
			default:
				keys := [][]byte{}
				ttlKeys := [][]byte{}

				err = ch.db.View(func(tx *bolt.Tx) error {
					c := tx.Bucket(bucketTTL).Cursor()

					max := []byte(time.Now().UTC().Format(time.RFC3339Nano))
					for k, v := c.First(); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
						keys = append(keys, v)
						ttlKeys = append(ttlKeys, k)
					}
					return nil
				})

				if err != nil {
					fmt.Println("error clean expired links", err)
				}

				err = ch.db.Update(func(tx *bolt.Tx) error {
					b := tx.Bucket(bucketTTL)
					for _, key := range ttlKeys {
						if err = b.Delete(key); err != nil {
							return err
						}
					}
					bitem := tx.Bucket(bucketData)
					for _, key := range keys {
						if err = bitem.Delete(key); err != nil {
							return err
						}
					}
					return nil
				})

				if err != nil {
					fmt.Println("error clean expired links", err)
				}
			}
			time.Sleep(time.Second)
		}
		ch.Done()
	}()

	return ch, nil
}

func (ch *BoltDBCache) Close() error {
	close(ch.stopCh)
	ch.Wait()
	return ch.db.Close()
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

func (ch *BoltDBCache) StoreExp(key interface{}, value interface{}, expiration int) {
	err := ch.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(cacheBucketName))
		err := b.Put([]byte(key.(string)), []byte(value.(string)))
		if err != nil {
			return err
		}
		bucketTTL := tx.Bucket([]byte(cacheBucketName + ":ttl"))
		expirationTime := time.Now().Add(time.Second * time.Duration(expiration)).UTC().Format(time.RFC3339Nano)
		err = bucketTTL.Put([]byte(expirationTime), []byte(key.(string)))
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
