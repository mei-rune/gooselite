package gooselite

import (
	"database/sql"
	"errors"
	"flag"
	"io"
	"io/fs"
	"os"
)

// Cmd is the interface each command must implement.
type Cmd interface {
	Flags(*flag.FlagSet) *flag.FlagSet
	Run(args []string) error
}

type commandEntry struct {
	Cmd
	Name    string
	Summary string
	Help    string
}

type Provider struct {
	cfg          *DBConfig
	migrationsFS fs.FS
	dialect      SqlDialect
	conn         *sql.DB
	args         map[string]string

	closers []io.Closer
}

// Option is a function that configures a Provider.
type Option func(*Provider)

// WithMigrationsFS sets the filesystem for migration files.
func WithTableName(tablename string) Option {
	return func(p *Provider) {
		p.cfg.TableName = tablename
	}
}

// WithArgs sets the command line arguments.
func WithArgs(args map[string]string) Option {
	return func(p *Provider) {
		p.args = args
	}
}

// WithMigrationsFS sets the filesystem for migration files.
func WithMigrationsFS(fsys fs.FS) Option {
	return func(p *Provider) {
		p.migrationsFS = fsys
	}
}

// WithMigrationsFS sets the filesystem for migration files.
func WithDir(dir string) Option {
	return func(p *Provider) {
		p.cfg.MigrationsDir = dir
	}
}

// WithDialect sets the SQL dialect.
func WithDialect(dialect SqlDialect) Option {
	return func(p *Provider) {
		p.dialect = dialect
	}
}

// WithConn sets the database connection.
func WithConn(conn *sql.DB) Option {
	return func(p *Provider) {
		p.conn = conn
	}
}

func NewProvider(cfg *DBConfig, opts ...Option) (*Provider, error) {
	p := &Provider{
		cfg:     cfg,
		dialect: DialectByName(cfg.DriverName),
	}
	for _, opt := range opts {
		opt(p)
	}
	if p.cfg.TableName == "" {
		p.cfg.TableName = "goose_db_version"
	} else {
		if err := ValidateTableName(p.cfg.TableName); err != nil {
			return nil, err
		}
	}

	if p.args == nil {
		p.args = make(map[string]string)
	}
	p.SetArg("driverName", p.cfg.DriverName)
	p.SetArg("tableName", p.cfg.GetTableName())
	return p, nil
}

func (p *Provider) Close() error {
	var errList error
	for _, c := range p.closers {
		if err := c.Close(); err != nil {
			if errList == nil {
				errList = err
			} else {
				errList = errors.Join(errList, err)
			}
		}
	}
	return errList
}

func (p *Provider) Conn() (*sql.DB, error) {
	if p.conn == nil {
		db, err := sql.Open(p.cfg.DriverName, p.cfg.ConnStr)
		if err != nil {
			return nil, err
		}
		p.conn = db
		p.closers = append(p.closers, db)
	}
	return p.conn, nil
}

func (p *Provider) SetArg(key, value string) {
	p.args[key] = value
}

func (p *Provider) GetMigrationsFS() fs.FS {
	if p.migrationsFS == nil {
		p.migrationsFS = os.DirFS(p.cfg.MigrationsDir)
	}
	return p.migrationsFS
}

func (p *Provider) GetTableName() string {
	return p.cfg.GetTableName()
}

// Dialect returns the SQL dialect used by this provider.
func (p *Provider) Dialect() SqlDialect {
	return p.dialect
}
