package bolt

import (
	"testing"
	"os"
	"fmt"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"encoding/binary"
	"enum-dns/enum"
)

var benchResult interface{}

func printRanges(number, enumID []byte) error {
	decodedNumber := binary.BigEndian.Uint64(number)
	decodedID := binary.BigEndian.Uint64(enumID)
	fmt.Printf("%d -> %d\n", decodedNumber, decodedID)
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

func Test_get_all(t *testing.T) {
	file, backend, err := createBoltBackend(t)
	if err != nil {
		t.FailNow()
	}
	defer os.Remove(file.Name())
	defer backend.Close()

	for i := 1; i <= 100; i++ {
		if _, err := backend.AddRange(enum.NumberRange{
			Lower:uint64(4740000000 + i * 10000),
			Upper:uint64(4740000000 + (i * 10000 - 1)),
			Regexp:fmt.Sprintf("range %d", i),
		}); err != nil {
			t.Error("Impossible to add range: ", err)
		}
	}

	ranges, err := backend.Ranges(4740049999, 10)
	if err != nil {
		t.Error("Impossible to get the ranges with backend.Ranges(1, 10): ", err)
		t.FailNow()
	}

	if len(ranges) != 10 {
		t.Error("Expected 10 ranges, got %d.", len(ranges))
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

	numberRange := enum.NumberRange{Lower:4741067196, Upper:4741067196, Regexp:"aregexp"}
	if _, err = backend.AddRange(numberRange); err != nil {
		t.Error("Error while saving range: ", err)
	}

	result, err := backend.RangeFor(41067196)
	if err != nil {
		t.Error("Error while retrieving the range: ", err)
	}

	if result.Regexp != numberRange.Regexp {
		t.Errorf("Expected %v, got %v", numberRange, result)
	}

}

func Test_create_overlap(t *testing.T) {

	file, backend, err := createBoltBackend(t)
	if err != nil {
		t.FailNow()
	}
	defer os.Remove(file.Name())
	defer backend.Close()

	numberRange := enum.NumberRange{Lower:4740000000, Upper:4749999999, Regexp:"mobile"}

	// Should fail
	failing := []enum.NumberRange{
		enum.NumberRange{Lower:4730000000, Upper:4740000000, Regexp:"fail"},
		enum.NumberRange{Lower:4749999999, Upper:4750000000, Regexp:"fail"},
		enum.NumberRange{Lower:4740000000, Upper:4750000000, Regexp:"fail"},
	}

	// Should work
	working := []enum.NumberRange{
		enum.NumberRange{Lower:4730000000, Upper:4739999999, Regexp:"aregexp"},
		enum.NumberRange{Lower:4750000000, Upper:4769999999, Regexp:"aregexp"},
	}

	if _, err = backend.AddRange(numberRange); err != nil {
		t.Error("Error while saving range: ", err)
	}

	for _, tt := range failing {
		if _, err = backend.AddRange(tt); err == nil {
			t.Error("Expected range conflict, but succeeded")
		}
	}

	for _, tt := range working {
		if _, err = backend.AddRange(tt); err != nil {
			t.Error("Expected to succeed, but failed", err)
		}
	}

	backend.Close()
	printBoltDatabase(file)
}

func TestNumber(t *testing.T) {

	var number uint64 = 4741067196
	var expectedResult uint64 = 474106719600000

	result, err := standardizeNumber(number)

	if err != nil {
		t.Errorf("Un expected error %v", err)
	} else {
		if result != expectedResult {
			t.Errorf("Expected %d, got %d", expectedResult, result)
		}
	}

}

func TestNumberZero(t *testing.T) {

	var number uint64 = 0
	var expectedResult uint64 = 0

	result, err := standardizeNumber(number)

	if err == nil {
		t.Errorf("Should have returned an error for %v", number)
	} else {
		if result != expectedResult {
			t.Errorf("Expected %d, got %d", expectedResult, result)
		}
	}

}

func TestNumberLimit(t *testing.T) {

	var number uint64 = 1000000000000000
	var expectedResult uint64 = 0

	result, err := standardizeNumber(number)

	if err == nil {
		t.Errorf("Should have returned an error for %v", number)
	} else {
		if result != expectedResult {
			t.Errorf("Expected %d, got %d", expectedResult, result)
		}
	}

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
		numberRange := enum.NumberRange{Lower:1000, Upper:1000, Regexp:""}
		backend.AddRange(numberRange)
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
		numberRange := enum.NumberRange{Lower:1000, Upper:2000, Regexp:""}
		backend.AddRange(numberRange)
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

	numberRange := enum.NumberRange{Lower:1000, Upper:2000, Regexp:""}
	backend.AddRange(numberRange)

	b.StartTimer()
	for i := 1; i < b.N; i++ {
		numberRange := enum.NumberRange{Lower:1500, Upper:2500, Regexp:""}
		backend.AddRange(numberRange)
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

	numberRange := enum.NumberRange{Lower:2000, Upper:3000, Regexp:""}
	backend.AddRange(numberRange)

	b.StartTimer()
	for i := 1; i < b.N; i++ {
		benchResult, _ = backend.RangeFor(1000)
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

	numberRange := enum.NumberRange{Lower:2000, Upper:3000, Regexp:""}
	backend.AddRange(numberRange)

	b.StartTimer()
	for i := 1; i < b.N; i++ {
		numberRange := enum.NumberRange{Lower:1500, Upper:2500, Regexp:""}
		backend.AddRange(numberRange)
	}
}
