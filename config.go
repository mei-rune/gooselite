package goose

import (
	"flag"
)

type DBConfig struct {
	MigrationsDir string
	DriverName    string
	ConnStr       string
}

// Flags registers the standard path/name/open flags on the given flag set.
func (c *DBConfig) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&c.MigrationsDir, "dir", "db", "migrations directory")
	fs.StringVar(&c.DriverName, "driver", "postgres", "driver name")
	fs.StringVar(&c.ConnStr, "conn", "", "driver connection string")
	return fs
}
