package goose

import (
	"context"
	"flag"
)

type DownCmd struct {
	cfg DBConfig
}

func (c *DownCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return c.cfg.Flags(fs)
}

func (c *DownCmd) Run(args []string) error {
	return Down(context.Background(), &c.cfg)
}

func Down(ctx context.Context, cfg *DBConfig) error {
	current, err := GetDBVersion(cfg)
	if err != nil {
		return err
	}

	previous, err := GetPreviousDBVersion(cfg.MigrationsDir, current)
	if err != nil {
		return err
	}

	return RunMigrations(ctx, cfg, cfg.MigrationsDir, previous)
}
