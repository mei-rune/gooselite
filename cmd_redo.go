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
	p, err := NewProvider(&c.cfg)
	if err != nil {
		return err
	}
	defer p.Close()
	return p.Redo(context.Background())
}

func (p *Provider) Redo(ctx context.Context) error {
	current, err := p.GetDBVersion(ctx)
	if err != nil {
		return err
	}

	previous, err := GetPreviousDBVersion(p.GetMigrationsFS(), current)
	if err != nil {
		return err
	}

	if err := p.RunMigrations(ctx, previous); err != nil {
		return err
	}

	return p.RunMigrations(ctx, current)
}
