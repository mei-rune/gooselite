package goose

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"
)

// global options. available to any subcommands.
var GooseFlagSet = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
var flagPath *string
var flagEnv *string
var flagTag *string

var ReadDbConf func() (*DBConf, error)

// helper to create a DBConf from the given flags
func dbConfFromFlags() (dbconf *DBConf, err error) {
	if nil != ReadDbConf {
		return ReadDbConf()
	}
	return NewDBConf(*flagPath, *flagEnv, *flagTag)
}

var Commands = []*Command{
	upCmd,
	downCmd,
	redoCmd,
	statusCmd,
	createCmd,
	dbVersionCmd,
	cleanCmd,
	resetCmd,
}

func Run(arguments ...string) {
	if nil == ReadDbConf && nil == GooseFlagSet.Lookup("path") {
		flagPath = GooseFlagSet.String("path", "db", "folder containing db info")
		flagEnv = GooseFlagSet.String("env", "development", "which DB environment to use")
		flagTag = GooseFlagSet.String("tag", "", "which config path and migrations path to use")
	}

	GooseFlagSet.Usage = usage
	GooseFlagSet.Parse(arguments)

	args := GooseFlagSet.Args()
	if len(args) == 0 || args[0] == "-h" {
		GooseFlagSet.Usage()
		return
	}

	var cmd *Command
	name := args[0]
	for _, c := range Commands {
		if strings.HasPrefix(c.Name, name) {
			cmd = c
			break
		}
	}

	if cmd == nil {
		fmt.Printf("error: unknown command %q\n", name)
		GooseFlagSet.Usage()
		os.Exit(1)
	}

	cmd.Exec(args[1:])
}

func usage() {
	fmt.Print(usagePrefix)
	GooseFlagSet.PrintDefaults()
	usageTmpl.Execute(os.Stdout, Commands)
}

var usagePrefix = `
goose is a database migration management system for Go projects.

Usage:
    goose [options] <subcommand> [subcommand options]

Options:
`
var usageTmpl = template.Must(template.New("usage").Parse(
	`
Commands:{{range .}}
    {{.Name | printf "%-10s"}} {{.Summary}}{{end}}
`))
