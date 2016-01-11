package bolt

import (
	"github.com/boltdb/bolt"
	"math"
	"errors"
	"encoding/binary"
	"encoding/json"
	"bytes"
	. "enum-dns/enum"
)

var (
	rangeBucket = []byte("range")
	enumBucket = []byte("enum")
)

type boltbackend struct {
	database *bolt.DB
}

func NewBoltDBBackend(fileName string) (Backend, error) {
	db, err := bolt.Open(fileName, 0600, nil)
	if err != nil {
		return nil, err
	}
	db.Update(func(tx *bolt.Tx) error {
		_, enumErr := tx.CreateBucketIfNotExists(enumBucket)
		if enumErr != nil {
			return enumErr
		}
		_, rangeErr := tx.CreateBucketIfNotExists(rangeBucket)
		if rangeErr != nil {
			return rangeErr
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return boltbackend{database:db}, err
}

func (b boltbackend) Close() error {
	return b.database.Close()
}

// Make the number 15 digits long
func standardizeNumber(number uint64) (uint64, error) {
	// Standardize the input.
	if !(0 < number && number < 1000000000000000) {
		return 0, errors.New("Number is outside the range [1:10^15]")
	}

	// 1234 -> 123400000000000 (E164).
	number = uint64(float64(number) * math.Pow10(int(14 - math.Floor(math.Log10(float64(number))))))

	return number, nil

}

// Transform uint64 to bytes
func Uint64ToBytes(input uint64) []byte {
	bytenum := make([]byte, 8)
	binary.BigEndian.PutUint64(bytenum, input)
	return bytenum
}

func (b boltbackend) Ranges(n uint64, c int) ([]NumberRange, error) {

	if c == 0 {
		return []NumberRange{}, nil
	}

	n, err := standardizeNumber(n)
	if err != nil {
		return nil, err
	}

	ranges := make([]NumberRange, 0, 20)
	err = b.database.View(func(tx *bolt.Tx) error {

		cur := tx.Bucket(rangeBucket).Cursor()

		Next := func(cur *bolt.Cursor, count *int) ([]byte, []byte) {
			if *count > 0 {
				*count = *count - 1
				cur.Next()
				return cur.Next()
			} else {
				*count = *count + 1
				cur.Prev()
				return cur.Prev()
			}
		}

		Check := func(current []byte, check []byte, count *int) bool {
			if *count == 0 {
				return false
			}
			if *count > 0 {
				return bytes.Compare(current, check) >= 0
			} else {
				return bytes.Compare(current, check) <= 0
			}
		}

		nb := Uint64ToBytes(n)
		for value, id := cur.Seek(nb); value != nil && Check(value, nb, &c); value, id = Next(cur, &c) {
			enumValue := tx.Bucket(enumBucket).Get(id)
			if enumValue == nil {
				return DBConsistencyError
			}

			var numRange NumberRange
			if err := json.Unmarshal(enumValue, &numRange); err != nil {
				return err
			}
			ranges = append(ranges, numRange)

		}
		return nil
	})

	return ranges, err

}

func (b boltbackend) RangeFor(number uint64) (numRange NumberRange, err error) {

	// Standardize the input.
	if number, err = standardizeNumber(number); err != nil {
		return
	}

	err = b.database.View(func(tx *bolt.Tx) error {

		cursor := tx.Bucket(rangeBucket).Cursor()
		key, value := cursor.Seek(Uint64ToBytes(number))
		if key != nil {
			enumValue := tx.Bucket(enumBucket).Get(value)
			err = json.Unmarshal(enumValue, &numRange)
		}
		return err
	})

	return
}

var (
	DBConsistencyError = errors.New("database inconsistent")
	ConflictingRangesError = errors.New("conflicting ranges")
)

func (b boltbackend) AddRange(r NumberRange) ([]NumberRange, error) {
	var err error
	var min, max []byte

	enumBytes, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	if stdMin, err := standardizeNumber(r.Lower); err != nil {
		return nil, err
	} else {
		min = Uint64ToBytes(stdMin)
	}

	if stdMax, err := standardizeNumber(r.Upper); err != nil {
		return nil, err
	} else {
		max = Uint64ToBytes(stdMax)
	}

	var conflictingRanges []NumberRange

	err = b.database.Update(func(tx *bolt.Tx) error {
		//fmt.Printf("Looking for conflict [%X:%X]\n", min, max)
		if bytes.Equal(min, max) {
			if enumID := tx.Bucket(rangeBucket).Get(min); enumID != nil {
				enumBytes := tx.Bucket(enumBucket).Get(enumID)
				if enumBytes == nil {
					return DBConsistencyError
				}

				var r NumberRange
				if err := json.Unmarshal(enumBytes, &r); err != nil {
					return err
				}

				conflictingRanges = []NumberRange{r}
			}
		} else {
			rbc := tx.Bucket(rangeBucket).Cursor()
			for n, ID := rbc.Seek(min); n != nil && bytes.Compare(n, max) <= 0; n, ID = rbc.Next() {
				//fmt.Printf("%d<%d<%d -> %X (%t)\n", binary.BigEndian.Uint64(min), binary.BigEndian.Uint64(n), binary.BigEndian.Uint64(max), ID, bytes.Compare(n, max) <= 0)
				enumBytes := tx.Bucket(enumBucket).Get(ID)
				if enumBytes == nil {
					return DBConsistencyError
				}
				var r NumberRange
				if err := json.Unmarshal(enumBytes, &r); err != nil {
					return err
				}
				conflictingRanges = append(conflictingRanges, r)
			}
		}
		//fmt.Printf("found conflicts:%v\n", conflictingRanges)
		if len(conflictingRanges) > 0 {
			return ConflictingRangesError
		} else {
			eb := tx.Bucket(enumBucket)

			enumID, _ := eb.NextSequence()
			rb := tx.Bucket(rangeBucket)
			if err := rb.Put(min, Uint64ToBytes(enumID)); err != nil {
				return err
			}
			if err := rb.Put(max, Uint64ToBytes(enumID)); err != nil {
				return err
			}
			if err := eb.Put(Uint64ToBytes(enumID), enumBytes); err != nil {
				return err
			}

			return nil
		}
	})

	if err == ConflictingRangesError {
		return conflictingRanges, err
	}

	return nil, err
}

func (b boltbackend) RemoveRange(r NumberRange) error {

	var min, max []byte
	if stdMin, err := standardizeNumber(r.Lower); err != nil {
		return err
	} else {
		min = Uint64ToBytes(stdMin)
	}

	if stdMax, err := standardizeNumber(r.Lower); err != nil {
		return err
	} else {
		max = Uint64ToBytes(stdMax)
	}

	err := b.database.Update(func(tx *bolt.Tx) error {
		rb := tx.Bucket(rangeBucket)
		minID := rb.Get(min)
		maxID := rb.Get(max)
		if minID == nil || maxID == nil || !bytes.Equal(minID, maxID) {
			return DBConsistencyError
		}
		eb := tx.Bucket(enumBucket)
		if enumID := eb.Get(minID); enumID == nil {
			return DBConsistencyError
		} else {
			if err := rb.Delete(minID); err != nil {
				return err
			}
			if err := rb.Delete(maxID); err != nil {
				return err
			}
			if err := eb.Delete(enumID); err != nil {
				return err
			}
		}

		return nil

	})

	return err
}