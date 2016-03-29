package mysql

import (
	"enum-dns/enum"
	"fmt"
	"testing"
)

func Test_UpdateIntervalQuery(t *testing.T) {

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
