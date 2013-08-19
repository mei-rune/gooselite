package goose

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"
)

var GooseFlagSet = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

func Run(arguments ...string) {

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
	usageTmpl.Execute(os.Stdout, Commands)
	fmt.Println("\noptions:")
	GooseFlagSet.PrintDefaults()
}

var usageTmpl = template.Must(template.New("usage").Parse(
	`goose is a database migration management system for Go projects.

Usage:
    goose [options] <subcommand> [subcommand options]

Commands:{{range .}}
    {{.Name | printf "%-10s"}} {{.Summary}}{{end}}
`))
