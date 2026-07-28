package gooselite

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
	p, err := NewProvider(&c.cfg)
	if err != nil {
		return err
	}
	defer p.Close()
	return p.Up(context.Background())
}

func (p *Provider) Up(ctx context.Context) error {
	target, err := GetMostRecentDBVersion(p.GetMigrationsFS())
	if err != nil {
		return err
	}

	return p.RunMigrations(ctx, target)
}
