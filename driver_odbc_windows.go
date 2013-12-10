package goose

import (
	//_ "code.google.com/p/odbc"
	_ "github.com/runner-mei/odbc"
)

func createMsSqlDriver(name, open string) DBDriver {
	return DBDriver{Name: "odbc",
		OpenStr: open,
		Import:  "code.google.com/p/odbc",
		Dialect: DialectByName("mssql")}
}

func init() {
	DBDrivers["mssql"] = createMsSqlDriver
}
