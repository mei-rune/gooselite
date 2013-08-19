package goose

import (
	_ "github.com/go-sql-driver/mysql"
)

func createMySqlDriver(name, open string) DBDriver {
	return DBDriver{Name: name,
		OpenStr: open,
		Import:  "github.com/go-sql-driver/mysql",
		Dialect: DialectByName("mysql")}
}

func init() {
	DBDrivers["mysql"] = createMySqlDriver
}
