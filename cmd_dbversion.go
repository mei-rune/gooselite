package gooselite

import (
	"context"
	"flag"
	"log"
)

type DbVersionCmd struct {
	cfg DBConfig
}

func (c *DbVersionCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return c.cfg.Flags(fs)
}

func (c *DbVersionCmd) Run(args []string) error {
	p, err := NewProvider(&c.cfg)
	if err != nil {
		return err
	}
	defer p.Close()
	return p.DbVersion(context.Background())
}

func (p *Provider) DbVersion(ctx context.Context) error {
	current, err := p.GetDBVersion(ctx)
	if err != nil {
		return err
	}

	log.Printf("goose: dbversion %v\n", current)
	return nil
}
