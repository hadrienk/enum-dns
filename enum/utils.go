package enum

import (
	"strconv"
	"strings"
)

// Reverse returns its argument string reversed rune-wise left to right.
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// Convert a enum DNS request to unsigned integer
func ConvertEnumToInt(enum string) (uint64, error) {
	// Remove the points.
	enum = strings.Replace(enum, ".", "", -1)
	enum = Reverse(enum)
	return strconv.ParseUint(enum, 10, 64)
}
