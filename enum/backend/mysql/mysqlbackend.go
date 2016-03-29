// staticbackend.go
package mysql

import (
	"database/sql"
	. "enum-dns/enum"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"math"
)

type mysqlbackend struct {
	database *sql.DB
}

// Create a new mysql backend
func NewMysqlBackend(driver string, dataSourceName string) (Backend, error) {
	con, err := sql.Open(driver, dataSourceName)
	return mysqlbackend{database: con}, err
}

func (b mysqlbackend) Close() error {
	return b.database.Close()
}

func (b mysqlbackend) PushRange(r NumberRange) ([]NumberRange, error) {

	tx, err := b.database.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.Query("SELECT lower, upper FROM interval WHERE lower <= ? AND ? <= upper;",
		r.Lower, r.Lower,
	)

	for rows.Next() {
		interval := NumberRange{}
		err = rows.Scan(interval.Lower, interval.Upper)
		if err != nil {
			return nil, err
		}

		updateQuery, args, err := updateIntervalQuery(r, interval).ToSql()
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(updateQuery, args...)
		if err != nil {
			return nil, err
		}

	}

	_, err = tx.Exec("INSERT INTO interval (lower, upper) VALUES (?, ?);", r.Lower, r.Upper)
	if err != nil {
		return nil, err
	}

	for _, record := range r.Records {
		_, err := tx.Exec("INSERT INTO record(lower, upper, order, preference, flags, service, regexp, replacement) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			r.Lower, r.Upper, record.Order,
			record.Preference, record.Flags,
			record.Service, record.Regexp, record.Replacement,
		)
		if err != nil {
			return nil, err
		}
	}

	tx.Commit()

	return nil, nil
}

func updateIntervalQuery(r NumberRange, n NumberRange) sq.Sqlizer {

	condition := sq.And{
		sq.Eq{"lower": r.Lower},
		sq.Eq{"upper": r.Upper},
	}
	if !n.Contains(r) {
		update := sq.Update("\"interval\"")
		switch {
		case r.Precedes(n):
			update.Set("upper", n.Lower-1)
		case r.Succeeds(n):
			update.Set("lower", n.Upper+1)
		}
		return update.Where(condition)
	} else {
		return sq.Delete("\"interval\"").Where(condition)
	}
}

func buildIntervalQueryFor(r NumberRange) sq.SelectBuilder {

	columns := []string{
		"i.lower", "i.upper",
		"r.order", "r.preference",
		"r.flags", "r.service",
		"r.regexp", "r.replacement",
	}
	interval := sq.Select(columns...).From("\"interval\" i")
	interval = interval.LeftJoin("record r ON r.lower = i.lower AND r.upper = i.upper")
	interval = interval.Where(
		sq.Or{
			sq.Expr("? BETWEEN i.lower AND i.upper", r.Lower),
			sq.Expr("? BETWEEN i.lower AND i.upper", r.Upper),
			sq.Expr("? <= i.lower AND ? >= i.upper", r.Lower, r.Upper),
		},
	)
	return interval
}

func (b mysqlbackend) RangesBetween(l, u uint64, c int) ([]NumberRange, error) {

	interval := buildIntervalQueryFor(NumberRange{Lower: l, Upper: u})
	switch {
	case c < 0:
		interval = interval.OrderBy("lower DESC")
		interval = interval.Limit(uint64(math.Abs(float64(c))))
	case c > 0:
		interval = interval.OrderBy("lower ASC")
		interval = interval.Limit(uint64(math.Abs(float64(c))))
	default:
		interval = interval.OrderBy("lower ASC")
	}

	query, args, err := interval.ToSql()
	if err != nil {
		return nil, err
	}
	fmt.Printf(query + "\n")
	rows, err := b.database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []NumberRange{}
	for rows.Next() {

		newRecord := Record{}
		newRange := NumberRange{}
		err := rows.Scan(
			&newRange.Lower, &newRange.Upper,
			&newRecord.Order, &newRecord.Preference,
			&newRecord.Flags, &newRecord.Service,
			&newRecord.Regexp, &newRecord.Replacement,
		)
		if err != nil {
			return nil, err
		}

		if len(results) == 0 {
			results = append(results, NumberRange{
				Lower: newRange.Lower, Upper: newRange.Upper,
				Records: []Record{newRecord},
			})
		} else {
			lastRange := &results[len(results)-1]
			if lastRange.Equals(newRange) {
				lastRange.Records = append(lastRange.Records, newRecord)
			} else {
				results = append(results, NumberRange{
					Lower: newRange.Lower, Upper: newRange.Upper,
					Records: []Record{newRecord},
				})
			}
		}
	}

	return results, err
}
