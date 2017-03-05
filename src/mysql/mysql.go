package mysql

import (
	"config"
	"fmt"
	"database/sql"
	"github.com/pkg/errors"
)

type Database struct {
	DB *sql.DB
}

type dbInfo struct {
	driverName string
	host       string
	port       int
	dbName     string
	username   string
	password   string
}

func (dm dbInfo) String() string {
	return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", dm.username, dm.password, dm.host, dm.port, dm.dbName)
}

func (database *Database) Connect() error {
	dm := dbInfo{}
	dm.driverName = config.DB_DRIVER
	dm.host = config.DB_HOST
	dm.port = config.DB_PORT
	dm.dbName = config.DB_NAME
	dm.username = config.DB_USERNAME
	dm.password = config.DB_PASSWORD

	var err error
	if database.DB, err = sql.Open(dm.driverName, dm.String()); err != nil {
		err = errors.Wrap(err, "Failed to connect to database.")
	}

	return err
}

func (database *Database) Close() error {
	err := database.DB.Close()

	return err
}

func (database *Database) QueryRow(sql string, args ...interface{}) *sql.Row {
	row := database.DB.QueryRow(sql, args...)

	return row
}

func (database *Database) Query(sql string, args ...interface{}) (*sql.Rows, error) {
	rows, err := database.DB.Query(sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to execute %v with %v", sql, args))
	}

	return rows, nil
}

func (database *Database) Execute(sql string, args ...interface{}) (sql.Result, error) {
	result, err := database.DB.Exec(sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to execute %v with %v", sql, args))
	}

	return result, nil
}

func (database *Database) FetchAll(sql string, args ...interface{}) ([][]interface{}, error) {
	result := make([][]interface{}, 0)

	rows, err := database.DB.Query(sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to execute %v with %v", sql, args))
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to get columns when executing %v with %v", sql, args))
	}

	length := len(cols)
	valAddr := make([]interface{}, length)

	for rows.Next() {
		vals := make([]interface{}, length)
		for i := range cols {
			valAddr[i] = &vals[i]
		}
		if err = rows.Scan(valAddr...); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Failed to get columns when executing %v with %v", sql, args))
		}

		for i := range cols {
			bs, ok := vals[i].([]byte)
			if ok {
				vals[i] = string(bs)
			}
		}

		result = append(result, vals)
	}

	return result, nil
}
