// staticbackend.go
package mysql

import (
	"database/sql"
	. "enum-dns/enum"
	"fmt"
)

type mysqlbackend struct {
	database *sql.DB
}

// Create a new mysql backend
func NewMysqlBackend(driver string, dataSourceName string) (Backend, error) {
	con, err := sql.Open("mysql", "enum:j8v6xkaK@tcp(127.0.0.1:3307)/enum")
	return mysqlbackend{database: con}, err
}

func (b mysqlbackend) Close() error {
	return b.database.Close()
}

func (b mysqlbackend) PushRange(r NumberRange) ([]NumberRange, error) {
	return nil, nil
}

func (b mysqlbackend) RangesBetween(l, u uint64, c int) (NumberRange, error) {

	queryTemplate := `SELECT i.lower,
							i.upper,
							r.order,
							r.preference,
							r.flags,
							r.service,
							r.regexp,
							r.replacement
						FROM interval i INNER JOIN records r
						ON r.lower = i.lower AND r.upper = i.upper
						WHERE ? <= i.lower  AND i.upper <= ? LIMIT ?;`

	rows, err := b.database.Query(queryTemplate, l, u, c)
	if err != nil {
		return err
	}
	defer rows.Close()

	results := make([]NumberRange, c)
	lastRange := new(NumberRange)
	for rows.Next() {
		record := new(Record)
		newRange := new(NumberRange)
		err := rows.Scan(
			newRange.Lower, newRange.Lower,
			record.Order, record.Preference,
			record.Flags, record.Service,
			record.Regexp, record.Replacement,
		)
		if err != nil {
			return nil, err
		}

		if !lastRange.Equals(newRange) {
			lastRange = newRange
			newRange.Records = make([]Record, 10)
			append(results, newRange)
		}
		append(lastRange.Records, record)
	}

	return results, err
}
