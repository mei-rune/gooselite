package goose

import (
	_ "github.com/alexbrainman/odbc"
)

func createMsSqlDriver(name, open string) DBDriver {
	return DBDriver{Name: "odbc",
		OpenStr: open,
		Import:  "github.com/alexbrainman/odbc",
		Dialect: DialectByName("mssql")}
}

func init() {
	DBDrivers["mssql"] = createMsSqlDriver
}
