package main

import (
	"bitbucket.org/liamstask/goose"
	_ "github.com/go-sql-driver/mysql"
)

func createMySqlDriver(name, open string) goose.DBDriver {
	return goose.DBDriver{Name: name,
		OpenStr: open,
		Import:  "github.com/go-sql-driver/mysql",
		Dialect: goose.DialectByName("mysql")}
}

func init() {
	goose.DBDrivers["mysql"] = createMySqlDriver
}
