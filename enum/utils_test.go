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
	"testing"
)

func TestReverse(t *testing.T) {
	expected := "esrever ot gnirts A"
	reversedString := Reverse("A string to reverse")

	if reversedString != expected {
		t.Error("Expected ", expected, " got ", reversedString)
	}

}

func TestConvertEnumToInt(t *testing.T) {

	expected := uint64(4741067196)
	enumString, err := ConvertEnumToInt("6.9.1.7.6.0.1.4.7.4")
	if err != nil {
		t.Error("Unexpeted error ", err)
	}

	if enumString != expected {
		t.Error("Expected ", enumString, " got ", expected)
	}

}

func TestPrefixToE164(t *testing.T) {
	tt := []struct {
		in   uint64
		exp  uint64
		fail bool
	}{
		{1000000000000000, 0, true},
		{0, 0, true},
		{1, 100000000000000, false},
		{2, 200000000000000, false},
		{123456, 123456000000000, false},
	}
	for _, v := range tt {
		if result, err := PrefixToE164(v.in); err != nil != v.fail {
			t.Error("Unexpected error: ", err)
		} else {
			if result != v.exp {
				t.Errorf("Expected PrefixToE164(%d) to return %d, got %d", v.in, v.exp, result)
			}
		}

	}
}
