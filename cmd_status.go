package goose

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"path/filepath"
)

type StatusCmd struct {
	cfg DBConfig
}

func (c *StatusCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return c.cfg.Flags(fs)
}

func (c *StatusCmd) Run(args []string) error {
	p, err := NewProvider(&c.cfg)
	if err != nil {
		return err
	}
	defer p.Close()
	return p.Status(context.Background())
}

func (p *Provider) Status(ctx context.Context) error {
	current, err := p.GetDBVersion(ctx)
	if err != nil {
		return err
	}

	fsys := p.GetMigrationsFS()

	min, err := GetPreviousDBVersion(fsys, current)
	if err != nil {
		min = 0
	}

	max, err := GetMostRecentDBVersion(fsys)
	if err != nil {
		return err
	}

	migrations, e := CollectMigrations(ctx, fsys, min, max)
	if e != nil {
		return e
	}

	db, err := p.Conn()
	if err != nil {
		return err
	}

	for _, m := range migrations {
		printMigrationStatus(ctx, p.dialect, db, p.cfg.GetTableName(), m.Version, filepath.Base(m.Source))
	}

	return nil
}

func printMigrationStatus(ctx context.Context, dialect SqlDialect, db *sql.DB, tableName string, version int64, name string) {
	rec, err := dialect.GetMigration(ctx, db, tableName, version)

	if err == nil {
		applied := ""
		if rec.IsApplied {
			applied = "up"
		} else {
			applied = "down"
		}
		fmt.Printf("%-6v %-20v %v %v\n", version, name, rec.TStamp.Format("2006-01-02 15:04:05"), applied)
	} else {
		log.Println(err)
	}
}
