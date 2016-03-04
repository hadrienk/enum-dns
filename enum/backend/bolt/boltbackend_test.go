package bolt

import (
	"bytes"
	"encoding/binary"
	"enum-dns/enum"
	"fmt"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"os"
	"testing"
)

var benchResult interface{}

func printRanges(number, enumID []byte) error {
	decodedNumber := binary.BigEndian.Uint64(number)
	left := binary.BigEndian.Uint64(enumID[:8])
	right := binary.BigEndian.Uint64(enumID[8:])
	fmt.Printf("%d -> %X:%X\n", decodedNumber, left, right)
	return nil
}

func printEnums(k, v []byte) error {
	key := binary.BigEndian.Uint64(k)
	fmt.Printf("key=%d, value=%s\n", key, v)
	return nil
}

func printBoltDatabase(file *os.File) {
	if db, err := bolt.Open(file.Name(), 0600, nil); err == nil {
		defer db.Close()
		db.View(func(tx *bolt.Tx) error {
			// Assume bucket exists and has keys
			rb := tx.Bucket([]byte("range"))
			fmt.Println("Range bucket:")
			rb.ForEach(printRanges)

			eb := tx.Bucket([]byte("enum"))
			fmt.Println("Enum bucket:")
			eb.ForEach(printEnums)
			return nil
		})
	}
}

func createTempFile(t testing.TB) (*os.File, error) {
	file, err := ioutil.TempFile("", "gotestbolt")
	if err != nil {
		t.Error("Could not create the file", err)
	} else {
		file.Close()
	}
	return file, err
}

func createBoltBackend(t testing.TB) (file *os.File, backend enum.Backend, err error) {
	if file, err = createTempFile(t); err != nil {
		return nil, nil, err
	}
	if backend, err = NewBoltDBBackend(file.Name()); err != nil {
		t.Error("Could not create the database", err)
		return nil, nil, err
	}
	return
}

func createBoltDatabase(t testing.TB) (file *os.File, db *bolt.DB, err error) {

	file, backend, err := createBoltBackend(t)
	if err != nil {
		t.FailNow()
	}
	backend.Close()

	db, err = bolt.Open(file.Name(), 0600, nil)

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		err := bucket.Put(uint64tobyte(100000000000001), uint64tobyte(1))
		err = bucket.Put(uint64tobyte(100000000000002), uint64tobyte(2))
		err = bucket.Put(uint64tobyte(100000000000003), uint64tobyte(3))
		err = bucket.Put(uint64tobyte(100000000000004), uint64tobyte(4))
		err = bucket.Put(uint64tobyte(100000000000005), uint64tobyte(5))
		err = bucket.Put(uint64tobyte(100000000000006), uint64tobyte(5))
		return err
	})

	return

}

// Ensures that values are returned in ascending order
func Test_rangesAscending(t *testing.T) {

	file, db, err := createBoltDatabase(t)
	if err != nil {
		t.FailNow()
	}
	defer os.Remove(file.Name())
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {

		cur := tx.Bucket(bucketName).Cursor()
		keys, _ := rangesAscending(cur, uint64tobyte(100000000000002), uint64tobyte(100000000000004), 0)

		if len(keys) != 3 {
			t.Errorf("expected 3 results, got %d", len(keys))
		}

		if bytes.Compare(keys[0], uint64tobyte(100000000000002)) != 0 {
			t.Errorf("expected key %X, got %X", uint64tobyte(100000000000002), keys[0])
		}
		if bytes.Compare(keys[1], uint64tobyte(100000000000003)) != 0 {
			t.Errorf("expected key %X, got %X", uint64tobyte(100000000000003), keys[1])
		}
		if bytes.Compare(keys[2], uint64tobyte(100000000000004)) != 0 {
			t.Errorf("expected key %d, got %d", 100000000000004, bytetouint64(keys[2]))
		}
		return nil
	})

}

// Ensures that values are returned in descending order
func Test_rangesDescending(t *testing.T) {

	file, db, err := createBoltDatabase(t)
	if err != nil {
		t.FailNow()
	}
	defer os.Remove(file.Name())
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {

		cur := tx.Bucket(bucketName).Cursor()
		keys, _ := rangesDescending(cur, uint64tobyte(100000000000002), uint64tobyte(100000000000004), 0)

		if len(keys) != 3 {
			t.Errorf("expected 3 results, got %d", len(keys))
		}

		if bytes.Compare(keys[2], uint64tobyte(100000000000002)) != 0 {
			t.Errorf("expected key %X, got %X", uint64tobyte(100000000000002), keys[2])
		}
		if bytes.Compare(keys[1], uint64tobyte(100000000000003)) != 0 {
			t.Errorf("expected key %X, got %X", uint64tobyte(100000000000003), keys[1])
		}
		if bytes.Compare(keys[0], uint64tobyte(100000000000004)) != 0 {
			t.Errorf("expected key %d, got %d", 100000000000004, bytetouint64(keys[0]))
		}
		return nil
	})

}

func Test_get_all(t *testing.T) {
	file, backend, err := createBoltBackend(t)
	if err != nil {
		t.FailNow()
	}
	defer backend.Close()
	defer os.Remove(file.Name())

	for i := 1; i <= 100; i++ {
		if _, err := backend.PushRange(enum.NumberRange{
			Lower:  uint64(4740000000 + i*10000),
			Upper:  uint64(4740000000 + (i*10000 - 1)),
			Regexp: fmt.Sprintf("range %d", i),
		}); err != nil {
			t.Error("Impossible to add range: ", err)
		}
	}

	ranges, err := backend.RangesBetween(474000000, 4740100000, 10)
	if err != nil {
		t.Error("Impossible to get the ranges with backend.Ranges(1, 10): ", err)
		t.FailNow()
	}

	if len(ranges) != 10 {
		t.Errorf("Expected 10 ranges, got %d.", len(ranges))
	}
	for _, r := range ranges {
		t.Logf("%v\n", r)
	}

	ranges, err = backend.RangesBetween(474000000, 4740100000, 5)
	if err != nil {
		t.Error("Impossible to get the ranges with backend.Ranges(1, 5): ", err)
		t.FailNow()
	}

	if len(ranges) != 5 {
		t.Errorf("Expected 5 ranges, got %d.", len(ranges))
	}
	for _, r := range ranges {
		t.Logf("%v\n", r)
	}

}

func Test_create(t *testing.T) {

	//printBoltDatabase()
	//file, err := os.OpenFile("/home/hadrien/enum.bolt", os.O_CREATE|os.O_RDWR, 0660)

	file, backend, err := createBoltBackend(t)
	if err != nil {
		t.FailNow()
	}
	defer os.Remove(file.Name())
	defer backend.Close()

	numberRange := enum.NumberRange{Lower: 4741067196, Upper: 4741067196, Regexp: "aregexp"}
	if _, err = backend.PushRange(numberRange); err != nil {
		t.Error("Error while saving range: ", err)
	}

	result, err := backend.RangesBetween(4741067196, 4741067196, 1)
	if err != nil {
		t.Error("Error while retrieving the range: ", err)
	}

	t.Log(result)
	if len(result) > 0 {
		if result[0].Regexp != numberRange.Regexp {
			t.Errorf("Expected %v, got %v", numberRange, result)
		}
	}

}

func Test_create_overlap(t *testing.T) {

	file, backend, err := createBoltBackend(t)
	if err != nil {
		t.FailNow()
	}
	defer os.Remove(file.Name())
	defer backend.Close()

	tt := []struct {
		start []enum.NumberRange
		in    []enum.NumberRange
		out   []enum.NumberRange
	}{{
		start: []enum.NumberRange{
			enum.NumberRange{Lower: 100000000000000, Upper: 100000000000002, Regexp: "default"},
			enum.NumberRange{Lower: 100000000000003, Upper: 100000000000005, Regexp: "default"},
			enum.NumberRange{Lower: 100000000000006, Upper: 100000000000008, Regexp: "default"},
		},
		in: []enum.NumberRange{
			enum.NumberRange{Lower: 100000000000001, Upper: 100000000000007, Regexp: "added"},
		},
		out: []enum.NumberRange{
			enum.NumberRange{Lower: 100000000000000, Upper: 100000000000000, Regexp: "default"},
			enum.NumberRange{Lower: 100000000000001, Upper: 100000000000007, Regexp: "default"},
			enum.NumberRange{Lower: 100000000000008, Upper: 100000000000008, Regexp: "default"},
		},
	}, {
		start: []enum.NumberRange{
			enum.NumberRange{Lower: 100000000000000, Upper: 100000000000002, Regexp: "default"},
			enum.NumberRange{Lower: 100000000000003, Upper: 100000000000005, Regexp: "default"},
			enum.NumberRange{Lower: 100000000000006, Upper: 100000000000008, Regexp: "default"},
		},
		in: []enum.NumberRange{
			enum.NumberRange{Lower: 100000000000002, Upper: 100000000000007, Regexp: "added"},
		},
		out: []enum.NumberRange{
			enum.NumberRange{Lower: 100000000000000, Upper: 100000000000001, Regexp: "default"},
			enum.NumberRange{Lower: 100000000000002, Upper: 100000000000007, Regexp: "default"},
			enum.NumberRange{Lower: 100000000000008, Upper: 100000000000008, Regexp: "default"},
		},
	}}

	for _, tt := range tt {

		for _, r := range tt.start {
			if _, err := backend.PushRange(r); err != nil {
				t.Fatalf("Could not insert test data %+v", r, err)
			}
		}

		for _, r := range tt.in {
			if _, err := backend.PushRange(r); err != nil {
				t.Fatalf("Could not insert range %+v, %s", r, err)
			}
		}

		results, err := backend.RangesBetween(100000000000000, 999999999999999, 10)
		if err != nil {
			t.Fatalf("Could not get the ranges")
		}

		if len(results) != len(tt.out) {
			t.Fatalf("Result size incorrect")
		}

		for i, r := range results {
			t.Logf("[%d:%d](%d)\n", r.Lower, r.Upper, i)
			if tt.out[i].Upper != r.Upper || tt.out[i].Lower != r.Lower {
				t.Errorf("Expected [%d:%d], got [%d:%d]", tt.out[i].Lower, tt.out[i].Upper, r.Lower, r.Upper)
			}
		}

		backend.Close()
		os.Remove(file.Name())
		file, backend, err = createBoltBackend(t)
		if err != nil {
			t.FailNow()
		}

	}

	backend.Close()
	//printBoltDatabase(file)
}

func BenchmarkAddSmall(b *testing.B) {
	b.StopTimer()

	file, backend, err := createBoltBackend(b)
	if err != nil {
		b.FailNow()
	}
	defer os.Remove(file.Name())
	defer backend.Close()

	b.StartTimer()
	for i := 1; i < b.N; i++ {
		numberRange := enum.NumberRange{Lower: 1000, Upper: 1000, Regexp: ""}
		backend.PushRange(numberRange)
	}

}

func BenchmarkAddEmpty(b *testing.B) {
	b.StopTimer()

	file, backend, err := createBoltBackend(b)
	if err != nil {
		b.FailNow()
	}
	defer os.Remove(file.Name())
	defer backend.Close()

	b.StartTimer()
	for i := 1; i < b.N; i++ {
		numberRange := enum.NumberRange{Lower: 1000, Upper: 2000, Regexp: ""}
		backend.PushRange(numberRange)
	}

}

func BenchmarkAddWithDupUp(b *testing.B) {
	b.StopTimer()

	file, backend, err := createBoltBackend(b)
	if err != nil {
		b.FailNow()
	}
	defer os.Remove(file.Name())
	defer backend.Close()

	numberRange := enum.NumberRange{Lower: 1000, Upper: 2000, Regexp: ""}
	backend.PushRange(numberRange)

	b.StartTimer()
	for i := 1; i < b.N; i++ {
		numberRange := enum.NumberRange{Lower: 1500, Upper: 2500, Regexp: ""}
		backend.PushRange(numberRange)
	}
}

func BenchmarkRangeFor(b *testing.B) {
	b.StopTimer()

	file, backend, err := createBoltBackend(b)
	if err != nil {
		b.FailNow()
	}
	defer os.Remove(file.Name())
	defer backend.Close()

	numberRange := enum.NumberRange{Lower: 2000, Upper: 3000, Regexp: ""}
	backend.PushRange(numberRange)

	b.StartTimer()
	for i := 1; i < b.N; i++ {
		benchResult, _ = backend.RangesBetween(1000, 1000, 1)
	}

}

func BenchmarkAddWithDupDown(b *testing.B) {
	b.StopTimer()

	file, backend, err := createBoltBackend(b)
	if err != nil {
		b.FailNow()
	}
	defer os.Remove(file.Name())
	defer backend.Close()

	numberRange := enum.NumberRange{Lower: 2000, Upper: 3000, Regexp: ""}
	backend.PushRange(numberRange)

	b.StartTimer()
	for i := 1; i < b.N; i++ {
		numberRange := enum.NumberRange{Lower: 1500, Upper: 2500, Regexp: ""}
		backend.PushRange(numberRange)
	}
}
