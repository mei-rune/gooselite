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
	p, err := NewProvider(&c.cfg)
	if err != nil {
		return err
	}
	defer p.Close()
	return p.Down(context.Background())
}

func (p *Provider) Down(ctx context.Context) error {
	current, err := p.GetDBVersion(ctx)
	if err != nil {
		return err
	}

	previous, err := GetPreviousDBVersion(p.GetMigrationsFS(), current)
	if err != nil {
		return err
	}

	return p.RunMigrations(ctx, previous)
}
