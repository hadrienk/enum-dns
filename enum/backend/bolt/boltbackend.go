package bolt

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	. "enum-dns/enum"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"math"
)

// The bolt bucket name
var bucketName []byte = []byte("range")

// The endianness to use.
var byteOrder binary.ByteOrder = binary.BigEndian

type boltbackend struct {
	database *bolt.DB
}

// Creates a new bolt backend. If the file does not exist it will be
// created automatically. The returned backend implements the io.Closer
// interface.
func NewBoltDBBackend(fileName string) (Backend, error) {
	db, err := bolt.Open(fileName, 0600, nil)
	if err != nil {
		return nil, err
	}
	db.Update(func(tx *bolt.Tx) error {
		_, enumErr := tx.CreateBucketIfNotExists(bucketName)
		if enumErr != nil {
			return enumErr
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return boltbackend{database: db}, err
}

func (b boltbackend) Close() error {
	return b.database.Close()
}

func rangesAscending(cur *bolt.Cursor, l, u []byte, limit uint) (keys [][]byte, values [][]byte) {
	keys = make([][]byte, 0)
	values = make([][]byte, 0)
	var k, v []byte
	for k, v = cur.Seek(l); k != nil && bytes.Compare(k, u) < 0; k, v = cur.Next() {
		keys = append(keys, k)
		values = append(values, v)
		limit = limit - 1
		if limit == 1 {
			// hit the limit
			return
		}
	}

	// Get the last one if not equal
	if k != nil && bytes.Compare(k, u) >= 0 {
		keys = append(keys, k)
		values = append(values, v)
	}
	return
}

func rangesDescending(cur *bolt.Cursor, l, u []byte, limit uint) (keys [][]byte, values [][]byte) {
	keys = make([][]byte, 0)
	values = make([][]byte, 0)
	k, v := cur.Seek(u)
	for ; k != nil && bytes.Compare(k, l) > 0; k, v = cur.Prev() {
		keys = append(keys, k)
		values = append(values, v)
		limit = limit - 1
		if limit == 1 {
			// hit the limit
			return
		}
	}

	// Get the last one if not equal
	if k != nil && bytes.Compare(k, l) <= 0 {
		keys = append(keys, k)
		values = append(values, v)
	}
	return
}

func rangesBetween(bucket *bolt.Bucket, l, u uint64, c int) (keys [][]byte, values [][]byte) {

	cur := bucket.Cursor()

	switch {
	case c < 0:
		return rangesDescending(cur, uint64tobyte(l), uint64tobyte(u), uint(math.Abs(float64(-1*c+1))))
	case c > 0:
		return rangesAscending(cur, uint64tobyte(l), uint64tobyte(u), uint(math.Abs(float64(1*c+1))))
	default:
		return rangesAscending(cur, uint64tobyte(l), uint64tobyte(u), 0)
	}

}

func (b boltbackend) RangesBetween(l, u uint64, c int) ([]NumberRange, error) {
	rangesToReturn := make([]NumberRange, 0, 20)

	if c == 0 {
		return rangesToReturn, nil
	}

	l, _ = standardizeNumber(l)
	u, _ = standardizeNumber(u)

	err := b.database.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket(bucketName)
		_, values := rangesBetween(bucket, l, u, c)
		for _, bytes := range values {
			numRange := NumberRange{}
			if err := json.Unmarshal(bytes, &numRange); err != nil {
				return err
			}
			rangesToReturn = append(rangesToReturn, numRange)
		}
		return nil

	})

	return rangesToReturn, err

}

var (
	DBConsistencyError     = errors.New("database inconsistent")
	ConflictingRangesError = errors.New("conflicting ranges")
)

// Tries to unmarshal from bucket. Fails if not data or format error.
func load(bucket *bolt.Bucket, key []byte) (NumberRange, error) {
	r := NumberRange{}
	bytes := bucket.Get(key)
	if bytes == nil {
		return r, errors.New(fmt.Sprintf("could not find data for the key %x", key))
	}

	if err := json.Unmarshal(bytes, &r); err != nil {
		return r, err
	}

	return r, nil
}

// Tries to marshal in bucket.
func save(bucket *bolt.Bucket, r NumberRange) error {
	bytes, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return bucket.Put(uint64tobyte(r.Upper), bytes)
}

func pushBefore(bucket *bolt.Bucket, key []byte, r NumberRange) error {
	toUpdate, err := load(bucket, key)
	if err != nil {
		return err
	}
	toUpdate.Upper = r.Lower - 1
	if err := bucket.Delete(key); err != nil {
		return err
	}

	return save(bucket, toUpdate)
}

func pushAfter(bucket *bolt.Bucket, key []byte, r NumberRange) error {
	toUpdate, err := load(bucket, key)
	if err != nil {
		return err
	}
	toUpdate.Lower = r.Upper + 1
	if err := bucket.Delete(key); err != nil {
		return err
	}

	return save(bucket, toUpdate)
}

func (b boltbackend) PushRange(r NumberRange) ([]NumberRange, error) {
	var err error

	r.Upper, err = standardizeNumber(r.Upper)
	if err != nil {
		return nil, err
	}

	r.Lower, err = standardizeNumber(r.Lower)
	if err != nil {
		return nil, err
	}

	//var conflictingRanges [][]byte

	err = b.database.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket(bucketName)
		keys, values := rangesBetween(bucket, r.Lower, r.Upper, 0)

		if len(keys) != len(values) {
			return DBConsistencyError
		}

		if len(keys) >= 1 {
			if err := pushBefore(bucket, keys[0], r); err != nil {
				return err
			}
		}

		if len(keys) >= 2 {
			if err := pushAfter(bucket, keys[len(keys)-1], r); err != nil {
				return err
			}
		}

		if len(keys) >= 3 {
			for _, key := range keys[1 : len(keys)-1] {
				if err := bucket.Delete(key); err != nil {
					return err
				}
			}
		}

		return save(bucket, r)

	})

	return nil, err
}
