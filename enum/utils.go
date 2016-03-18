// Copyright 2016 Hadrien Kohl hadrien.kohl@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package enum

import (
	"errors"
	"math"
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

// Make any number 15 digits long by padding zeros
func PrefixToE164(number uint64) (uint64, error) {
	// Standardize the input.
	if !(0 < number && number < 1000000000000000) {
		return 0, errors.New("Number is outside the range [1:10^15]")
	}

	// 1234 -> 123400000000000 (E164).
	number = uint64(float64(number) * math.Pow10(int(14-math.Floor(math.Log10(float64(number))))))

	return number, nil

}
