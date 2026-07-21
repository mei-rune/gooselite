package goose

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"
)

type StatusCmd struct {
	cfg DBConfig
}

type StatusData struct {
	Source string
	Status string
}

func (c *StatusCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return c.cfg.Flags(fs)
}

func (c *StatusCmd) Run(args []string) error {
	return Status(context.Background(), &c.cfg)
}

func Status(ctx context.Context, cfg *DBConfig) error {
	// collect all migrations
	min := int64(0)
	max := int64((1 << 63) - 1)
	migrations, e := CollectMigrations(ctx, cfg.MigrationsDir, min, max)
	if e != nil {
		return e
	}

	db, e := sql.Open(cfg.DriverName, cfg.ConnStr)
	if e != nil {
		return fmt.Errorf("couldn't open DB: %w", e)
	}
	defer db.Close()

	dialect := DialectByName(cfg.DriverName)

	// must ensure that the version table exists if we're running on a pristine DB
	if _, e := EnsureDBVersion(ctx, cfg, db); e != nil {
		return e
	}

	fmt.Println("    Applied At                  Migration")
	fmt.Println("    =======================================")
	for _, m := range migrations {
		printMigrationStatus(ctx, dialect, db, m.Version, filepath.Base(m.Source))
	}
	return nil
}

func printMigrationStatus(ctx context.Context, dialect SqlDialect, db *sql.DB, version int64, script string) {
	row, e := dialect.GetMigration(ctx, db, version)

	// q := fmt.Sprintf("SELECT tstamp, is_applied FROM goose_db_version WHERE version_id=? ORDER BY tstamp DESC LIMIT 1", version)
	// e := db.QueryRow(q).Scan(&row.TStamp, &row.IsApplied)

	if e != nil && e != sql.ErrNoRows {
		log.Fatal(e)
	}

	var appliedAt string

	if row.IsApplied {
		appliedAt = row.TStamp.Format(time.ANSIC)
	} else {
		appliedAt = "Pending"
	}

	fmt.Printf("    %-24s -- %v\n", appliedAt, script)
}
