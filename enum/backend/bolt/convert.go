package bolt

import (
	"errors"
	"math"
)

func bytetouint64(b []byte) uint64 {
	return byteOrder.Uint64(b)
}

func uint64tobyte(i uint64) []byte {
	bytes := make([]byte, 8)
	byteOrder.PutUint64(bytes, i)
	return bytes
}

// Make the number 15 digits long
func standardizeNumber(number uint64) (uint64, error) {
	// Standardize the input.
	if !(0 < number && number < 1000000000000000) {
		return 0, errors.New("Number is outside the range [1:10^15]")
	}

	// 1234 -> 123400000000000 (E164).
	number = uint64(float64(number) * math.Pow10(int(14-math.Floor(math.Log10(float64(number))))))

	return number, nil

}
