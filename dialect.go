package gooselite

import (
	"context"
	"database/sql"
	"fmt"
)

type DBRunner interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// SqlDialect abstracts the details of specific SQL dialects
// for goose's few SQL specific statements
type SqlDialect interface {
	CreateVersionTableSql(ctx context.Context, db DBRunner, tableName string) error                                                 // create the goose_db_version table
	InsertVersionSql(ctx context.Context, db DBRunner, tableName string, versionId int64, direction bool, description string) error // insert the initial version table row
	DbVersionQuery(ctx context.Context, db DBRunner, tableName string) (*sql.Rows, error)
	GetMigration(ctx context.Context, db DBRunner, tableName string, versionId int64) (MigrationRecord, error)
}

type GenDialect struct {
	CreateTableSQL string
	InsertSQL      string
	QuerySQL       string
	FetchSQL       string
}

func (d *GenDialect) CreateVersionTableSql(ctx context.Context, db DBRunner, tableName string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf(d.CreateTableSQL, tableName))
	return err
}

func (d *GenDialect) InsertVersionSql(ctx context.Context, db DBRunner, tableName string, versionId int64, direction bool, description string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf(d.InsertSQL, tableName), versionId, direction, description)
	return err
}

func (d *GenDialect) DbVersionQuery(ctx context.Context, db DBRunner, tableName string) (*sql.Rows, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(d.QuerySQL, tableName))
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (d *GenDialect) GetMigration(ctx context.Context, db DBRunner, tableName string, versionId int64) (MigrationRecord, error) {
	var row MigrationRecord
	e := db.QueryRowContext(ctx, fmt.Sprintf(d.FetchSQL, tableName), versionId).Scan(&row.TStamp, &row.IsApplied, &row.Description)
	if e != nil {
		return MigrationRecord{}, e
	}
	return row, nil
}

type CreateSqlDialect func() SqlDialect

var (
	SqlDialects = map[string]CreateSqlDialect{
		"postgres": func() SqlDialect {
			return &GenDialect{
				CreateTableSQL: `CREATE TABLE %s (
					id serial NOT NULL,
					version_id bigint NOT NULL,
					is_applied boolean NOT NULL,
					description text,
					tstamp timestamp NULL default now(),
					PRIMARY KEY(id)
				);`,
				InsertSQL: "INSERT INTO %s (version_id, is_applied, description) VALUES ($1, $2, $3);",
				QuerySQL:  "SELECT version_id, is_applied, description from %s ORDER BY id DESC",
				FetchSQL:  "SELECT tstamp, is_applied, description FROM %s WHERE version_id=$1 ORDER BY tstamp DESC LIMIT 1",
			}
		},
		"mysql": func() SqlDialect {
			return &GenDialect{
				CreateTableSQL: `CREATE TABLE %s (
					id serial NOT NULL,
					version_id bigint NOT NULL,
					is_applied boolean NOT NULL,
					description text,
					tstamp timestamp NULL default now(),
					PRIMARY KEY(id)
				);`,
				InsertSQL: "INSERT INTO %s (version_id, is_applied, description) VALUES (?, ?, ?);",
				QuerySQL:  "SELECT version_id, is_applied, description from %s ORDER BY id DESC",
				FetchSQL:  "SELECT tstamp, is_applied, description FROM %s WHERE version_id=? ORDER BY tstamp DESC LIMIT 1",
			}
		},
		"sqlite": func() SqlDialect {
			return &GenDialect{
				CreateTableSQL: `CREATE TABLE %s (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					version_id INTEGER NOT NULL,
					is_applied INTEGER NOT NULL,
					description text,
					tstamp TIMESTAMP DEFAULT (datetime('now'))
				);`,
				InsertSQL: "INSERT INTO %s (version_id, is_applied, description) VALUES (?, ?, ?);",
				QuerySQL:  "SELECT version_id, is_applied, description from %s ORDER BY id DESC",
				FetchSQL:  "SELECT tstamp, is_applied, description FROM %s WHERE version_id=? ORDER BY tstamp DESC LIMIT 1",
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
