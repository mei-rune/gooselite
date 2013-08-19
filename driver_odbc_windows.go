package goose

import (
	_ "code.google.com/p/odbc"
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
