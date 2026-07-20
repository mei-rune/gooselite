package goose

import (
	"context"
	"flag"
	"fmt"
)

type DbVersionCmd struct {
	cfg DBConfig
}

func (c *DbVersionCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return c.cfg.Flags(fs)
}

func (c *DbVersionCmd) Run(args []string) error {
	return DbVersion(context.Background(), &c.cfg)
}

func DbVersion(ctx context.Context, cfg *DBConfig) error {
	current, err := GetDBVersion(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("goose: dbversion %v\n", current)
	return nil
}
