package goose

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/lib/pq"

	_ "gitee.com/opengauss/openGauss-connector-go-pq" // openGauss
	_ "gitee.com/chunanyong/dm"                       // 达梦
	_ "gitee.com/runner.mei/gokb"                     // 人大金仓
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/sijms/go-ora/v2" // oracle
)

// DBDriver encapsulates the info needed to work with
// a specific database driver
type DBDriver struct {
	Name    string
	OpenStr string
	Import  string
	Dialect SqlDialect
}

type DBConf struct {
	MigrationsDir string
	Env           string
	Driver        DBDriver
}

// extract configuration details from the given file
func NewDBConf(path, env, tag string) (*DBConf, error) {
	cfgFile := filepath.Join(path, "dbconf.yml")
	if 0 != len(tag) {
		file := filepath.Join(path, "dbconf-"+tag+".yml")
		if st, e := os.Stat(file); nil == e && nil != st && !st.IsDir() {
			cfgFile = file
		}
	}

	f, err := yaml.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}

	drv, err := f.Get(fmt.Sprintf("%s.driver", env))
	if err != nil {
		return nil, err
	}

	open, err := f.Get(fmt.Sprintf("%s.open", env))
	if err != nil {
		return nil, err
	}
	open = os.ExpandEnv(open)

	// Automatically parse postgres urls
	if drv == "postgres" {

		// Assumption: If we can parse the URL, we should
		if parsedURL, err := pq.ParseURL(open); err == nil && parsedURL != "" {
			open = parsedURL
		}
	}

	d := NewDBDriver(drv, open)

	// allow the configuration to override the Import for this driver
	if imprt, err := f.Get(fmt.Sprintf("%s.import", env)); err == nil {
		d.Import = imprt
	}

	// allow the configuration to override the Dialect for this driver
	if dialect, err := f.Get(fmt.Sprintf("%s.dialect", env)); err == nil {
		d.Dialect = DialectByName(dialect)
	}

	if !d.IsValid() {
		return nil, errors.New(fmt.Sprintf("Invalid DBConf: %v", d))
	}

	migrations_path := filepath.Join(path, "migrations")
	if 0 != len(tag) {
		pa := filepath.Join(path, "migrations-"+tag)
		if st, e := os.Stat(pa); nil == e && nil != st && st.IsDir() {
			migrations_path = pa
		}
	}

	return &DBConf{
		MigrationsDir: migrations_path,
		Env:           env,
		Driver:        d,
	}, nil
}

type CreateDBDriver func(name, open string) DBDriver

var (
	DBDrivers = map[string]CreateDBDriver{}
)

// Create a new DBDriver and populate driver specific
// fields for drivers that we know about.
// Further customization may be done in NewDBConf
func NewDBDriver(name, open string) DBDriver {
	if createDBDriver, ok := DBDrivers[name]; ok {
		return createDBDriver(name, open)
	}
	return DBDriver{Name: name,
		OpenStr: open}
}

// ensure we have enough info about this driver
func (drv *DBDriver) IsValid() bool {
	return len(drv.Import) > 0 && drv.Dialect != nil
}

func createPostgresDriver(name, open string) DBDriver {
	return DBDriver{Name: name,
		OpenStr: open,
		Import:  "github.com/lib/pq",
		Dialect: DialectByName("postgres")}
}

func createMyMySqlDriver(name, open string) DBDriver {
	return DBDriver{Name: name,
		OpenStr: open,
		Import:  "github.com/ziutek/mymysql/godrv",
		Dialect: DialectByName("mysql")}
}

func createSqlite3Driver(name, open string) DBDriver {
	return DBDriver{
		Name:    name,
		OpenStr: open,
		Import:  "github.com/mattn/go-sqlite3",
		Dialect: DialectByName("sqlite"),
	}
}

func init() {
	DBDrivers["postgres"] = createPostgresDriver
	DBDrivers["mymysql"] = createMyMySqlDriver
	DBDrivers["sqlite3"] = createSqlite3Driver
}
