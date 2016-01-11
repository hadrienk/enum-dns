package enum

import "testing"

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