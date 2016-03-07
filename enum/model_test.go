package enum

import "testing"

func Test_OverlapsWith(t *testing.T) {

	r := NumberRange{Lower: 400000000000000, Upper: 500000000000000}

	tt := []struct {
		r   NumberRange
		exp bool
	}{
		{NumberRange{Lower: 400000000000000, Upper: 500000000000000}, true},
		{NumberRange{Lower: 399999999999999, Upper: 500000000000000}, true},
		{NumberRange{Lower: 400000000000000, Upper: 500000000000001}, true},
		{NumberRange{Lower: 399999999999999, Upper: 500000000000001}, true},

		{NumberRange{Lower: 500000000000001, Upper: 500000000000002}, false},
		{NumberRange{Lower: 399999999999997, Upper: 399999999999998}, false},
	}

	for _, v := range tt {
		if r.OverlapWith(v.r) != v.exp {
			t.Error("[%d:%d].OverlapWith([%d:%d]) returned %b, expected %b",
				r.Lower, r.Upper,
				v.r.Lower, v.r.Upper,
				r.OverlapWith(v.r), v.exp,
			)
		}
	}

}
