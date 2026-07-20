package goose

import (
	"context"
	"flag"
)

type RedoCmd struct {
	cfg DBConfig
}

func (c *RedoCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return c.cfg.Flags(fs)
}

func (c *RedoCmd) Run(args []string) error {
	return Redo(context.Background(), &c.cfg)
}

func Redo(ctx context.Context, cfg *DBConfig) error {
	current, err := GetDBVersion(cfg)
	if err != nil {
		return err
	}

	previous, err := GetPreviousDBVersion(cfg.MigrationsDir, current)
	if err != nil {
		return err
	}

	if err := RunMigrations(ctx, cfg, cfg.MigrationsDir, previous); err != nil {
		return err
	}

	return RunMigrations(ctx, cfg, cfg.MigrationsDir, current)
}
