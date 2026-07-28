package gooselite

import (
	"context"
	"flag"
	"log"
)

type CleanCmd struct {
	cfg DBConfig
}

func (c *CleanCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return c.cfg.Flags(fs)
}

func (c *CleanCmd) Run(args []string) error {
	p, err := NewProvider(&c.cfg)
	if err != nil {
		return err
	}
	defer p.Close()
	return p.Clean(context.Background())
}

func (p *Provider) Clean(ctx context.Context) error {
	current, err := p.GetDBVersion(ctx)
	if err != nil {
		return err
	}

	if current == 0 {
		log.Println("db is empty, can't clean.")
		return nil
	}

	return p.RunMigrations(ctx, 0)
}
