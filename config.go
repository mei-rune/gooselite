package gooselite

import (
	"flag"
)

const DefaultTableName = "goose_db_version"

type DBConfig struct {
	MigrationsDir string
	TableName     string
	DriverName    string
	ConnStr       string
}

func (c *DBConfig) GetTableName() string {
	if c.TableName == "" {
		return DefaultTableName
	}
	return c.TableName
}

// Flags registers the standard path/name/open flags on the given flag set.
func (c *DBConfig) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&c.MigrationsDir, "dir", "db", "migrations directory")
	fs.StringVar(&c.TableName, "table", DefaultTableName, "table name")
	fs.StringVar(&c.DriverName, "driver", "postgres", "driver name")
	fs.StringVar(&c.ConnStr, "conn", "", "driver connection string")
	return fs
}
