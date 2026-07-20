package goose

import (
	"context"
	"flag"
	"fmt"
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

	return Create(context.Background(), c.migrationsDir, args[0], migrationType)
}

func Create(ctx context.Context, migrationsDir, name, migrationType string) error {
	if err := os.MkdirAll(migrationsDir, 0777); err != nil {
		return err
	}

	n, err := CreateMigration(name, migrationType, migrationsDir, time.Now())
	if err != nil {
		return err
	}

	a, e := filepath.Abs(n)
	if e != nil {
		return e
	}

	fmt.Println("goose: created", a)
	return nil
}
