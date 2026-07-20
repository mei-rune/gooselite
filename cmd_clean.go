package goose

import (
	"context"
	"flag"
	"fmt"
)

type CleanCmd struct {
	cfg DBConfig
}

func (c *CleanCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return c.cfg.Flags(fs)
}

func (c *CleanCmd) Run(args []string) error {
	return Clean(context.Background(), &c.cfg)
}

func Clean(ctx context.Context, cfg *DBConfig) error {
	current, err := GetDBVersion(cfg)
	if err != nil {
		return err
	}

	if current == 0 {
		fmt.Println("db is empty, can't clean.")
		return nil
	}

	return RunMigrations(ctx, cfg, cfg.MigrationsDir, 0)
}
