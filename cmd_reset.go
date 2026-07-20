package goose

import (
	"context"
	"flag"
)

type ResetCmd struct {
	cfg DBConfig
}

func (c *ResetCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return c.cfg.Flags(fs)
}

func (c *ResetCmd) Run(args []string) error {
	return Reset(context.Background(), &c.cfg)
}

func Reset(ctx context.Context, cfg *DBConfig) error {
	current, err := GetDBVersion(cfg)
	if err != nil {
		return err
	}

	if current != 0 {
		if err := RunMigrations(ctx, cfg, cfg.MigrationsDir, 0); err != nil {
			return err
		}
	}

	target, err := GetMostRecentDBVersion(cfg.MigrationsDir)
	if err != nil {
		return err
	}
	return RunMigrations(ctx, cfg, cfg.MigrationsDir, target)
}
