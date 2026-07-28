package gooselite

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type CreateCmd struct {
	migrationsDir string
}

func (c *CreateCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&c.migrationsDir, "dir", "db", "migrations directory")
	return fs
}

func (c *CreateCmd) Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("goose create: migration name required")
	}

	migrationType := "go" // default to Go migrations
	if len(args) >= 2 {
		migrationType = args[1]
	}

	p, err := NewProvider(&DBConfig{MigrationsDir: c.migrationsDir})
	if err != nil {
		return err
	}
	defer p.Close()
	return p.Create(context.Background(), args[0], migrationType)
}

func (p *Provider) Create(ctx context.Context, name, migrationType string) error {
	if err := os.MkdirAll(p.cfg.MigrationsDir, 0777); err != nil {
		return err
	}

	n, err := CreateMigration(name, migrationType, p.cfg.MigrationsDir, time.Now())
	if err != nil {
		return err
	}

	a, e := filepath.Abs(n)
	if e != nil {
		return e
	}

	log.Println("goose: created", a)
	return nil
}
