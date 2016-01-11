// staticbackend.go
package mysql

import (
	"fmt"
	"database/sql"
	. "enum-dns/enum"
)

type mysqlbackend struct {
	database *sql.DB
}

// Create a new mysql backend
func NewMysqlBackend(driver string, dataSourceName string) (Backend, error) {
	con, err := sql.Open("mysql", "enum:j8v6xkaK@tcp(127.0.0.1:3307)/enum")
	return mysqlbackend{database:con}, err
}

func (b mysqlbackend) Close() error {
	return b.database.Close()
}

func (b mysqlbackend) RemoveRange(r NumberRange) error {
	return nil
}
func (b mysqlbackend) AddRange(r NumberRange) ([]NumberRange, error) {
	return nil, nil
}
func (b mysqlbackend) RangeFor(number uint64) (NumberRange, error) {
	row := b.database.QueryRow("select lower, upper, regex from enum where lower <= ? AND ? <= upper", number, number)
	result := new(NumberRange)
	err := row.Scan(&result.Lower, &result.Upper, &result.Regexp)
	if err != nil {
		fmt.Printf("[ERR] dns: database error: %v", err)
		return *result, err
	} else {
		return *result, nil
	}
}
