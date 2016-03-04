package bolt

import (
	"bytes"
	"testing"
)

func Test_uint64_to_bytes(t *testing.T) {
	if bytetouint64(uint64tobyte(654123)) != 654123 {
		t.Errorf("bytetouint64(uint64tobyte(654123)) => %d, want %d", bytetouint64(uint64tobyte(654123)), 654123)
	}
}

func Test_compare(t *testing.T) {
	if bytes.Compare(uint64tobyte(1), uint64tobyte(1)) != 0 {
		t.Errorf("bytes.Compare(uint64tobyte(1), uint64tobyte(1)) => %d, want %d", bytes.Compare(uint64tobyte(1), uint64tobyte(1)), 0)
	}
}

func Test_compare_sup(t *testing.T) {
	if bytes.Compare(uint64tobyte(1), uint64tobyte(2)) != -1 {
		t.Errorf("bytes.Compare(uint64tobyte(1), uint64tobyte(2)) => %d, want %d", bytes.Compare(uint64tobyte(1), uint64tobyte(1)), -1)
	}
}

func Test_compare_inf(t *testing.T) {
	if bytes.Compare(uint64tobyte(2), uint64tobyte(1)) != 1 {
		t.Errorf("bytes.Compare(uint64tobyte(2), uint64tobyte(1)) => %d, want %d", bytes.Compare(uint64tobyte(2), uint64tobyte(1)), 1)
	}
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
