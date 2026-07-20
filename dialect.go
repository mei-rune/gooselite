package goose

import (
	"context"
	"database/sql"
)

type DBRunner interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// SqlDialect abstracts the details of specific SQL dialects
// for goose's few SQL specific statements
type SqlDialect interface {
	CreateVersionTableSql(ctx context.Context, db DBRunner) error // create the goose_db_version table
	InsertVersionSql(ctx context.Context, db DBRunner) error      // insert the initial version table row
	DbVersionQuery(ctx context.Context, db DBRunner) (*sql.Rows, error)
}

type GenDialect struct {
	CreateTableSQL string
	InsertSQL      string
	QuerySQL       string
}

func (d *GenDialect) CreateVersionTableSql(ctx context.Context, db DBRunner) error {
	_, err := db.ExecContext(ctx, d.CreateTableSQL)
	return err
}

func (d *GenDialect) InsertVersionSql(ctx context.Context, db DBRunner) error {
	_, err := db.ExecContext(ctx, d.InsertSQL, 0, true)
	return err
}

func (d *GenDialect) DbVersionQuery(ctx context.Context, db DBRunner) (*sql.Rows, error) {
	rows, err := db.QueryContext(ctx, d.QuerySQL)
	if err != nil {
		return nil, ErrTableDoesNotExist
	}
	return rows, err
}

type CreateSqlDialect func() SqlDialect

var (
	SqlDialects = map[string]CreateSqlDialect{
		"postgres": func() SqlDialect {
			return &GenDialect{
				CreateTableSQL: `CREATE TABLE goose_db_version (
					id serial NOT NULL,
					version_id bigint NOT NULL,
					is_applied boolean NOT NULL,
					tstamp timestamp NULL default now(),
					PRIMARY KEY(id)
				);`,
				InsertSQL: "INSERT INTO goose_db_version (version_id, is_applied) VALUES ($1, $2);",
				QuerySQL:  "SELECT version_id, is_applied from goose_db_version ORDER BY id DESC",
			}
		},
		"mysql": func() SqlDialect {
			return &GenDialect{
				CreateTableSQL: `CREATE TABLE goose_db_version (
					id serial NOT NULL,
					version_id bigint NOT NULL,
					is_applied boolean NOT NULL,
					tstamp timestamp NULL default now(),
					PRIMARY KEY(id)
				);`,
				InsertSQL: "INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, ?);",
				QuerySQL:  "SELECT version_id, is_applied from goose_db_version ORDER BY id DESC",
			}
		},
		"sqlite": func() SqlDialect {
			return &GenDialect{
				CreateTableSQL: `CREATE TABLE goose_db_version (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					version_id INTEGER NOT NULL,
					is_applied INTEGER NOT NULL,
					tstamp TIMESTAMP DEFAULT (datetime('now'))
				);`,
				InsertSQL: "INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, ?);",
				QuerySQL:  "SELECT version_id, is_applied from goose_db_version ORDER BY id DESC",
			}
		},
	}
)

// drivers that we don't know about can ask for a dialect by name
func DialectByName(d string) SqlDialect {
	if createSqlDialect, ok := SqlDialects[d]; ok {
		return createSqlDialect()
	}
	return nil
}
