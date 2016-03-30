package mysql

import (
	"enum-dns/enum"
	"fmt"
	"testing"
)

func Test_UpdateIntervalQuery(t *testing.T) {

	state := enum.NumberRange{Lower: 4, Upper: 6}

	tt := []struct {
		in    enum.NumberRange
		query string
		args  []uint64
		err   error
	}{
		{
			enum.NumberRange{Lower: 3, Upper: 5},
			`UPDATE "interval" SET lower = ? WHERE (lower = ? AND upper = ?)`,
			[]uint64{6, 4, 6}, nil,
		}, {
			enum.NumberRange{Lower: 5, Upper: 7},
			`UPDATE "interval" SET upper = ? WHERE (lower = ? AND upper = ?)`,
			[]uint64{4, 4, 6}, nil,
		}, {
			enum.NumberRange{Lower: 1, Upper: 8},
			`DELETE FROM "interval" WHERE (lower = ? AND upper = ?)`, []uint64{4, 6}, nil,
		},
	}

	for _, v := range tt {
		t.Logf("Testing updateIntervalQuery(%v, %v)", state, v.in)
		query, args, err := updateIntervalQuery(state, v.in).ToSql()
		if err != v.err {
			t.Errorf("Unexpected error: ", err)
		}
		if query != v.query {
			t.Errorf("Incorrect query.\nExpected %s\nGot %s",
				v.query, query,
			)
		}
		if !(len(args) != len(v.args)) {
			for i, value := range args {
				if value != v.args[i] {
					t.Errorf("Incorrect args. Expected %v, got %v", v.args, args)
				}
			}
		}

	}

	//t.Log(updateIntervalQuery(state, enum.NumberRange{Lower: 3, Upper: 5}).ToSql())
	//t.Log(updateIntervalQuery(state, enum.NumberRange{Lower: 5, Upper: 7}).ToSql())
	//t.Log(updateIntervalQuery(state, enum.NumberRange{Lower: 1, Upper: 2}).ToSql())
	//t.Log(updateIntervalQuery(state, enum.NumberRange{Lower: 7, Upper: 8}).ToSql())
	//t.Log(updateIntervalQuery(state, enum.NumberRange{Lower: 3, Upper: 7}).ToSql())
}

func Test_OverlapingQuery(t *testing.T) {
	r := enum.NumberRange{Lower: 1, Upper: 9}

	er := "SELECT i.lower, i.upper, r.order, r.preference, r.flags, r.service, r.regexp, r.replacement FROM interval i LEFT JOIN record r ON r.lower = i.lower AND r.upper = i.upper WHERE (? BETWEEN i.lower AND i.upper OR ? BETWEEN i.lower AND i.upper OR ? <= i.lower AND ? >= i.upper)"

	result, args, err := buildIntervalQueryFor(r).ToSql()
	if err != nil {
		t.Fatal("Error: ", err)
	}

	if fmt.Sprintf(result, args...) != fmt.Sprintf(er, r.Lower, r.Upper, r.Lower, r.Upper) {
		t.Fatalf("Expected query to be %s, but got %s",
			fmt.Sprintf(er, r.Lower, r.Upper, r.Lower, r.Upper),
			fmt.Sprintf(result, args...),
		)
	}

}
