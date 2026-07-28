package goose

import (
	"context"
	"flag"
	"log"
)

type EnsureCmd struct {
	cfg DBConfig
}

func (c *EnsureCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return c.cfg.Flags(fs)
}

func (c *EnsureCmd) Run(args []string) error {
	p, err := NewProvider(&c.cfg)
	if err != nil {
		return err
	}
	defer p.Close()
	_, err = p.EnsureDBVersion(context.Background())
	return err
}

// retrieve the current version for this DB.
// Create and initialize the DB version table if it doesn't exist.
func (p *Provider) EnsureDBVersion(ctx context.Context) (int64, error) {
	db, err := p.Conn()
	if err != nil {
		return 0, err
	}

	rows, err := p.dialect.DbVersionQuery(ctx, db, p.cfg.GetTableName())
	if err != nil {
		if IsTableNotExists(err) {
			return 0, p.createVersionTable(ctx)
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
		if err = rows.Scan(&row.VersionId, &row.IsApplied, &row.Description); err != nil {
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
