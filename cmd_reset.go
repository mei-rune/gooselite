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
	p, err := NewProvider(&c.cfg)
	if err != nil {
		return err
	}
	defer p.Close()
	return p.Reset(context.Background())
}

func (p *Provider) Reset(ctx context.Context) error {
	current, err := p.GetDBVersion(ctx)
	if err != nil {
		return err
	}

	if current != 0 {
		if err := p.RunMigrations(ctx, 0); err != nil {
			return err
		}
	}

	target, err := GetMostRecentDBVersion(p.GetMigrationsFS())
	if err != nil {
		return err
	}
	return p.RunMigrations(ctx, target)
}
