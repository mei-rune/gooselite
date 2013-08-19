package main

import (
	"bitbucket.org/liamstask/goose"
	_ "code.google.com/p/odbc"
	"database/sql"
)

////////////////////////////
// MS SQLServer
////////////////////////////

type MSSQLDialect struct{}

func (pg *MSSQLDialect) CreateVersionTableSql() string {
	return `CREATE TABLE goose_db_version (
            	  id INT IDENTITY(1,1) NOT NULL,
                version_id bigint NOT NULL,
                is_applied BIT NOT NULL,
                tstamp timestamp,
                PRIMARY KEY(id)
            );`
}

func (pg *MSSQLDialect) InsertVersionSql() string {
	return "INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, ?);"
}

func (pg *MSSQLDialect) DbVersionQuery(db *sql.DB) (*sql.Rows, error) {
	rows, err := db.Query("SELECT version_id, is_applied from goose_db_version ORDER BY id DESC")

	// XXX: check for postgres specific error indicating the table doesn't exist.
	// for now, assume any error is because the table doesn't exist,
	// in which case we'll try to create it.
	if err != nil {
		return nil, goose.ErrTableDoesNotExist
	}

	return rows, err
}

func createMsSqlDriver(name, open string) goose.DBDriver {
	return goose.DBDriver{Name: "odbc",
		OpenStr: open,
		Import:  "github.com/go-sql-driver/mysql",
		Dialect: goose.DialectByName("mysql")}
}

func init() {
	goose.DBDrivers["mssql"] = createMsSqlDriver
}
