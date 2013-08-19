package main

import (
	"database/sql"
)

////////////////////////////
// MS SQLServer
////////////////////////////

type MSSQLDialect struct{}

func (pg *MSSQLDialect) createVersionTableSql() string {
	return `CREATE TABLE goose_db_version (
            	  id INT IDENTITY(1,1) NOT NULL,
                version_id bigint NOT NULL,
                is_applied BIT NOT NULL,
                tstamp timestamp,
                PRIMARY KEY(id)
            );`
}

func (pg *MSSQLDialect) insertVersionSql() string {
	return "INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, ?);"
}

func (pg *MSSQLDialect) dbVersionQuery(db *sql.DB) (*sql.Rows, error) {
	rows, err := db.Query("SELECT version_id, is_applied from goose_db_version ORDER BY id DESC")

	// XXX: check for postgres specific error indicating the table doesn't exist.
	// for now, assume any error is because the table doesn't exist,
	// in which case we'll try to create it.
	if err != nil {
		return nil, ErrTableDoesNotExist
	}

	return rows, err
}
