package goose

import (
	"database/sql"
)

// SqlDialect abstracts the details of specific SQL dialects
// for goose's few SQL specific statements
type SqlDialect interface {
	CreateVersionTableSql() string // sql string to create the goose_db_version table
	InsertVersionSql() string      // sql string to insert the initial version table row
	DbVersionQuery(db *sql.DB) (*sql.Rows, error)
}

type CreateSqlDialect func() SqlDialect

var (
	SqlDialects = map[string]CreateSqlDialect{"postgres": func() SqlDialect { return &PostgresDialect{} },
		"mysql":  func() SqlDialect { return &MySqlDialect{} },
		"sqlite": func() SqlDialect { return &SqliteDialect{} }}
)

// drivers that we don't know about can ask for a dialect by name
func DialectByName(d string) SqlDialect {
	if createSqlDialect, ok := SqlDialects[d]; ok {
		return createSqlDialect()
	}
	return nil
}

////////////////////////////
// Postgres
////////////////////////////

type PostgresDialect struct{}

func (pg *PostgresDialect) CreateVersionTableSql() string {
	return `CREATE TABLE goose_db_version (
            	id serial NOT NULL,
                version_id bigint NOT NULL,
                is_applied boolean NOT NULL,
                tstamp timestamp NULL default now(),
                PRIMARY KEY(id)
            );`
}

func (pg *PostgresDialect) InsertVersionSql() string {
	return "INSERT INTO goose_db_version (version_id, is_applied) VALUES ($1, $2);"
}

func (pg *PostgresDialect) DbVersionQuery(db *sql.DB) (*sql.Rows, error) {
	rows, err := db.Query("SELECT version_id, is_applied from goose_db_version ORDER BY id DESC")

	// XXX: check for postgres specific error indicating the table doesn't exist.
	// for now, assume any error is because the table doesn't exist,
	// in which case we'll try to create it.
	if err != nil {
		return nil, ErrTableDoesNotExist
	}

	return rows, err
}

////////////////////////////
// MySQL
////////////////////////////

type MySqlDialect struct{}

func (m *MySqlDialect) CreateVersionTableSql() string {
	return `CREATE TABLE goose_db_version (
                id serial NOT NULL,
                version_id bigint NOT NULL,
                is_applied boolean NOT NULL,
                tstamp timestamp NULL default now(),
                PRIMARY KEY(id)
            );`
}

func (m *MySqlDialect) InsertVersionSql() string {
	return "INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, ?);"
}

func (m *MySqlDialect) DbVersionQuery(db *sql.DB) (*sql.Rows, error) {
	rows, err := db.Query("SELECT version_id, is_applied from goose_db_version ORDER BY id DESC")

	// XXX: check for mysql specific error indicating the table doesn't exist.
	// for now, assume any error is because the table doesn't exist,
	// in which case we'll try to create it.
	if err != nil {
		return nil, ErrTableDoesNotExist
	}

	return rows, err
}

////////////////////////////
// Sqlite
////////////////////////////

type SqliteDialect struct{}

func (m *SqliteDialect) CreateVersionTableSql() string {
	return `CREATE TABLE goose_db_version (
                id integer primary key,
                version_id bigint NOT NULL,
                is_applied boolean NOT NULL,
                tstamp timestamp NULL default CURRENT_TIMESTAMP
            );`
}

func (m *SqliteDialect) InsertVersionSql() string {
	return "INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, ?);"
}

func (m *SqliteDialect) DbVersionQuery(db *sql.DB) (*sql.Rows, error) {
	rows, err := db.Query("SELECT version_id, is_applied from goose_db_version ORDER BY id DESC")

	// XXX: check for mysql specific error indicating the table doesn't exist.
	// for now, assume any error is because the table doesn't exist,
	// in which case we'll try to create it.
	if err != nil {
		return nil, ErrTableDoesNotExist
	}

	return rows, err
}
