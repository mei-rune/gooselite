package goose

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var (
	ErrTableDoesNotExist = errors.New("table does not exist")
	ErrNoPreviousVersion = errors.New("no previous version found")
)

type MigrationRecord struct {
	VersionId int64
	TStamp    time.Time
	IsApplied bool // was this a result of up() or down()
}

type Migration struct {
	Version  int64
	Next     int64  // next version, or -1 if none
	Previous int64  // previous version, -1 if none
	Source   string // path to .go or .sql script
}

type migrationSorter []*Migration

// helpers so we can use pkg sort
func (ms migrationSorter) Len() int           { return len(ms) }
func (ms migrationSorter) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }
func (ms migrationSorter) Less(i, j int) bool { return ms[i].Version < ms[j].Version }

func newMigration(v int64, src string) *Migration {
	return &Migration{v, -1, -1, src}
}

func RunMigrations(ctx context.Context, conf *DBConfig, migrationsDir string, target int64) (err error) {
	db, err := sql.Open(conf.DriverName, conf.ConnStr)
	if err != nil {
		return err
	}
	defer db.Close()

	current, err := EnsureDBVersion(ctx, conf, db)
	if err != nil {
		return err
	}

	migrations, err := CollectMigrations(ctx, migrationsDir, current, target)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		fmt.Printf("goose: no migrations to run. current version: %d\n", current)
		return nil
	}

	ms := migrationSorter(migrations)
	direction := current < target
	ms.Sort(direction)

	dialect := DialectByName(conf.DriverName)

	fmt.Printf("goose: migrating db, current version: %d, target: %d\n",
		current, target)

	for _, m := range ms {

		switch filepath.Ext(m.Source) {
		case ".sql":
			err = runSQLMigration(ctx, conf, dialect, db, m.Source, m.Version, direction)
		}

		if err != nil {
			return err
		}

		fmt.Println("OK   ", filepath.Base(m.Source))
	}

	return nil
}

// collect all the valid looking migration scripts in the
// migrations folder, and key them by version
func CollectMigrations(ctx context.Context, dirpath string, current, target int64) (m []*Migration, err error) {

	// extract the numeric component of each migration,
	// filter out any uninteresting files,
	// and ensure we only have one file per migration version.
	err = filepath.Walk(dirpath, func(name string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if v, e := NumericComponent(name); e == nil {

			for _, g := range m {
				if v == g.Version {
					return fmt.Errorf("more than one file specifies the migration for version %d (%s and %s)",
						v, g.Source, filepath.Join(dirpath, name))
				}
			}

			if versionFilter(v, current, target) {
				m = append(m, newMigration(v, name))
			}
		}

		return nil
	})

	return m, err
}

func versionFilter(v, current, target int64) bool {
	if target > current {
		return v > current && v <= target
	}

	if target < current {
		return v <= current && v > target
	}

	return false
}

func (ms migrationSorter) Sort(direction bool) {
	sort.Sort(ms)

	// reverse order if needed
	if direction == false {
		for i, j := 0, len(ms)-1; i < j; i, j = i+1, j-1 {
			ms[i], ms[j] = ms[j], ms[i]
		}
	}

	// now that we're sorted in the appropriate direction,
	// populate next and previous for each migration
	for i, m := range ms {
		prev := int64(-1)
		if i > 0 {
			prev = ms[i-1].Version
			ms[i-1].Next = m.Version
		}
		ms[i].Previous = prev
	}
}

// look for migration scripts with names in the form:
//
//	XXX_descriptivename.ext
//
// where XXX specifies the version number
// and ext specifies the type of migration
func NumericComponent(name string) (int64, error) {
	base := filepath.Base(name)

	if ext := filepath.Ext(base); ext != ".go" && ext != ".sql" {
		return 0, errors.New("not a recognized migration file type")
	}

	idx := strings.Index(base, "_")
	if idx < 0 {
		return 0, errors.New("no separator found")
	}

	n, e := strconv.ParseInt(base[:idx], 10, 64)
	if e == nil && n <= 0 {
		return 0, errors.New("migration IDs must be greater than zero")
	}

	return n, e
}

// retrieve the current version for this DB.
// Create and initialize the DB version table if it doesn't exist.
func EnsureDBVersion(ctx context.Context, conf *DBConfig, db *sql.DB) (int64, error) {
	rows, err := DialectByName(conf.DriverName).DbVersionQuery(ctx, db)
	if err != nil {

		if err == ErrTableDoesNotExist {
			return 0, createVersionTable(ctx, conf, db)
		}
		return 0, err
	}
	defer rows.Close()

	// The most recent record for each migration specifies
	// whether it has been applied or rolled back.
	// The first version we find that has been applied is the current version.

	toSkip := make([]int64, 0)

	for rows.Next() {
		var row MigrationRecord
		if err = rows.Scan(&row.VersionId, &row.IsApplied); err != nil {
			log.Fatal("error scanning rows:", err)
		}

		// have we already marked this version to be skipped?
		skip := false
		for _, v := range toSkip {
			if v == row.VersionId {
				skip = true
				break
			}
		}

		// if version has been applied and not marked to be skipped, we're done
		if row.IsApplied && !skip {
			return row.VersionId, nil
		}

		// version is either not applied, or we've already seen a more
		// recent version of it that was not applied.
		if !skip {
			toSkip = append(toSkip, row.VersionId)
		}
	}

	panic("failure in EnsureDBVersion()")
}

// Create the goose_db_version table
// and insert the initial 0 value into it
func createVersionTable(ctx context.Context, conf *DBConfig, db *sql.DB) error {
	txn, err := db.Begin()
	if err != nil {
		return fmt.Errorf("db.Begin: %w", err)
	}

	d := DialectByName(conf.DriverName)

	if err := d.CreateVersionTableSql(ctx, db); err != nil {
		txn.Rollback()
		if strings.Contains(err.Error(), "already exists") ||
			strings.Contains(err.Error(), "已存在") ||
			strings.Contains(err.Error(), "已经存在") ||
			strings.Contains(err.Error(), "ORA-00955") {
			return nil
		}
		return fmt.Errorf("create version table: %w", err)
	}

	if err := d.InsertVersionSql(ctx, db, 0, true); err != nil {
		txn.Rollback()
		return fmt.Errorf("insert initial version: %w", err)
	}

	return txn.Commit()
}

// wrapper for EnsureDBVersion for callers that don't already have
// their own DB instance
func GetDBVersion(conf *DBConfig) (version int64, err error) {
	db, err := sql.Open(conf.DriverName, conf.ConnStr)
	if err != nil {
		return -1, err
	}
	defer db.Close()

	version, err = EnsureDBVersion(context.Background(), conf, db)
	if err != nil {
		return -1, err
	}

	return version, nil
}

func GetPreviousDBVersion(dirpath string, version int64) (previous int64, err error) {
	previous = -1
	sawGivenVersion := false

	filepath.Walk(dirpath, func(name string, info os.FileInfo, walkerr error) error {

		if !info.IsDir() {
			if v, e := NumericComponent(name); e == nil {
				if v > previous && v < version {
					previous = v
				}
				if v == version {
					sawGivenVersion = true
				}
			}
		}

		return nil
	})

	if previous == -1 {
		if sawGivenVersion {
			// the given version is (likely) valid but we didn't find
			// anything before it.
			// 'previous' must reflect that no migrations have been applied.
			previous = 0
		} else {
			err = ErrNoPreviousVersion
		}
	}

	return
}

// helper to identify the most recent possible version
// within a folder of migration scripts
func GetMostRecentDBVersion(dirpath string) (version int64, err error) {
	version = -1

	e := filepath.Walk(dirpath, func(name string, info os.FileInfo, walkerr error) error {
		if nil != walkerr {
			return walkerr
		}
		if !info.IsDir() {
			if v, e := NumericComponent(name); e == nil {
				if v > version {
					version = v
				}
			}
		}

		return nil
	})

	if nil != e {
		err = errors.New("no valid version found in the '" + dirpath + "' - " + e.Error())
		return
	}

	if version == -1 {
		err = errors.New("no valid version found in the '" + dirpath + "'.")
	}

	return
}

func CreateMigration(name, migrationType, dir string, t time.Time) (path string, err error) {
	if migrationType != "go" && migrationType != "sql" {
		return "", errors.New("migration type must be 'go' or 'sql'")
	}

	timestamp := t.Format("20060102150405")
	filename := fmt.Sprintf("%v_%v.%v", timestamp, name, migrationType)

	fpath := filepath.Join(dir, filename)

	var tmpl *template.Template
	if migrationType == "sql" {
		tmpl = sqlMigrationTemplate
	} else {
		tmpl = goMigrationTemplate
	}

	path, err = writeTemplateToFile(fpath, tmpl, timestamp)

	return
}

var goMigrationTemplate = template.Must(template.New("goose.go-migration").Parse(`
package main

import (
	"database/sql"
)

// Up is executed when this migration is applied
func Up_{{ . }}(txn *sql.Tx) {

}

// Down is executed when this migration is rolled back
func Down_{{ . }}(txn *sql.Tx) {

}
`))

var sqlMigrationTemplate = template.Must(template.New("goose.sql-migration").Parse(`
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

`))
