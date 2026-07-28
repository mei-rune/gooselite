package gooselite

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var (
	ErrTableDoesNotExist     = errors.New("table does not exist")
	ErrTableDoesAlreadyExist = errors.New("table does already exist")
	ErrNoPreviousVersion     = errors.New("no previous version found")
)

type MigrationRecord struct {
	VersionId   int64
	TStamp      time.Time
	IsApplied   bool   // was this a result of up() or down()
	Description string // description of the migration
}

type Migration struct {
	Version     int64
	Next        int64  // next version, or -1 if none
	Previous    int64  // previous version, -1 if none
	Source      string // path to .go or .sql script
	Description string // description derived from filename
}

type migrationSorter []*Migration

// helpers so we can use pkg sort
func (ms migrationSorter) Len() int           { return len(ms) }
func (ms migrationSorter) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }
func (ms migrationSorter) Less(i, j int) bool { return ms[i].Version < ms[j].Version }

func newMigration(v int64, src string) *Migration {
	desc := filepath.Base(src)
	desc = strings.TrimSuffix(desc, filepath.Ext(desc))
	if idx := strings.Index(desc, "_"); idx >= 0 {
		desc = desc[idx+1:]
	}
	return &Migration{
		Version:     v,
		Next:        -1,
		Previous:    -1,
		Source:      src,
		Description: desc,
	}
}

func (p *Provider) RunMigrations(ctx context.Context, target int64) (err error) {
	db, err := p.Conn()
	if err != nil {
		return err
	}

	current, err := p.EnsureDBVersion(ctx)
	if err != nil {
		return err
	}

	fsys := p.GetMigrationsFS()

	migrations, err := CollectMigrations(ctx, fsys, current, target)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		log.Printf("goose: no migrations to run. current version: %d\n", current)
		return nil
	}

	ms := migrationSorter(migrations)
	direction := current < target
	ms.Sort(direction)

	log.Printf("goose: migrating db, current version: %d, target: %d\n",
		current, target)

	for _, m := range ms {
		args := map[string]interface{}{
			"version":     m.Version,
			"direction":   direction,
			"description": m.Description,
		}
		for key, value := range p.args {
			args[key] = value
		}

		switch strings.ToLower(filepath.Ext(m.Source)) {
		case ".sql":
			err = runSQLMigration(ctx, p.dialect, db, p.cfg.GetTableName(), fsys, m.Source, m.Version, direction, args)
		}

		if err != nil {
			return err
		}

		log.Println("OK   ", filepath.Base(m.Source))
	}

	return nil
}

// collect all the valid looking migration scripts in the
// migrations folder, and key them by version
func CollectMigrations(ctx context.Context, fsys fs.FS, current, target int64) (m []*Migration, err error) {

	// extract the numeric component of each migration,
	// filter out any uninteresting files,
	// and ensure we only have one file per migration version.
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if v, e := NumericComponent(path); e == nil {

			for _, g := range m {
				if v == g.Version {
					return fmt.Errorf("more than one file specifies the migration for version %d (%s and %s)",
						v, g.Source, path)
				}
			}

			if versionFilter(v, current, target) {
				m = append(m, newMigration(v, path))
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

// Create the goose_db_version table
// and insert the initial 0 value into it
func (p *Provider) createVersionTable(ctx context.Context) error {
	db, err := p.Conn()
	if err != nil {
		return err
	}

	txn, err := db.Begin()
	if err != nil {
		return fmt.Errorf("db.Begin: %w", err)
	}

	if err := p.dialect.CreateVersionTableSql(ctx, db, p.cfg.GetTableName()); err != nil {
		txn.Rollback()
		if IsTableAlreadyExists(err) {
			return nil
		}
		return fmt.Errorf("create version table: %w", err)
	}

	if err := p.dialect.InsertVersionSql(ctx, db, p.cfg.GetTableName(), 0, true, "init version"); err != nil {
		txn.Rollback()
		return fmt.Errorf("insert initial version: %w", err)
	}

	return txn.Commit()
}

// wrapper for EnsureDBVersion for callers that don't already have
// their own DB instance
func (p *Provider) GetDBVersion(ctx context.Context) (version int64, err error) {
	return p.EnsureDBVersion(ctx)
}

func GetPreviousDBVersion(fsys fs.FS, version int64) (previous int64, err error) {
	previous = -1
	sawGivenVersion := false

	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if v, e := NumericComponent(path); e == nil {
			if v > previous && v < version {
				previous = v
			}
			if v == version {
				sawGivenVersion = true
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
func GetMostRecentDBVersion(fsys fs.FS) (version int64, err error) {
	version = -1

	e := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if nil != err {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if v, e := NumericComponent(path); e == nil {
			if v > version {
				version = v
			}
		}

		return nil
	})

	if nil != e {
		err = errors.New("no valid version found - " + e.Error())
		return
	}

	if version == -1 {
		err = errors.New("no valid version found.")
	}

	return
}

func CreateMigration(name, migrationType, dir string, t time.Time) (path string, err error) {
	// if migrationType != "go" && migrationType != "sql" {
	//	return "", errors.New("migration type must be 'go' or 'sql'")
	// }
	if migrationType != "sql" {
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
