package goose

import (
	//_ "code.google.com/p/odbc"
	// _ "github.com/runner-mei/odbc"
	// _ "github.com/alexbrainman/odbc"
	_ "github.com/denisenkom/go-mssqldb"
)

func createMsSqlDriver(name, open string) DBDriver {
	return DBDriver{
		Name:    "mssql",
		OpenStr: open,
		Import:  "github.com/denisenkom/go-mssqldb",
		Dialect: DialectByName("mssql"),
	}
}

func init() {
	DBDrivers["mssql"] = createMsSqlDriver
}
