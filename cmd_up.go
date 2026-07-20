package goose

import (
	"context"
	"flag"
)

type UpCmd struct {
	cfg DBConfig
}

func (c *UpCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return c.cfg.Flags(fs)
}

func (c *UpCmd) Run(args []string) error {
	return Up(context.Background(), &c.cfg)
}

func Up(ctx context.Context, cfg *DBConfig) error {
	target, err := GetMostRecentDBVersion(cfg.MigrationsDir)
	if err != nil {
		return err
	}

	return RunMigrations(ctx, cfg, cfg.MigrationsDir, target)
}
