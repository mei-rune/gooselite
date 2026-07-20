package goose

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var Commands = []commandEntry{
	{Cmd: &UpCmd{}, Name: "up", Summary: "Migrate the DB to the most recent version available", Help: "up extended help here..."},
	{Cmd: &DownCmd{}, Name: "down", Summary: "Roll back the version by 1", Help: "down extended help here..."},
	{Cmd: &RedoCmd{}, Name: "redo", Summary: "Re-run the latest migration", Help: "redo extended help here..."},
	{Cmd: &StatusCmd{}, Name: "status", Summary: "dump the migration status for the current DB", Help: "status extended help here..."},
	{Cmd: &CreateCmd{}, Name: "create", Summary: "Create the scaffolding for a new migration", Help: "create extended help here..."},
	{Cmd: &DbVersionCmd{}, Name: "dbversion", Summary: "Print the current version of the database", Help: "dbversion extended help here..."},
	{Cmd: &CleanCmd{}, Name: "clean", Summary: "Roll back the version by 1", Help: "clean extended help here..."},
	{Cmd: &ResetCmd{}, Name: "reset", Summary: "Roll back the version to 0 and Migrate the DB to the most recent version available", Help: "reset extended help here..."},
}

func Run(arguments ...string) {
	if len(arguments) == 0 || arguments[0] == "-h" {
		fmt.Println("Available commands:")
		for _, c := range Commands {
			fmt.Printf("    %s\n", c.Name)
		}
		return
	}

	// first argument is the subcommand name
	name := arguments[0]

	var entry *commandEntry
	for i, c := range Commands {
		if strings.HasPrefix(c.Name, name) {
			entry = &Commands[i]
			break
		}
	}

	if entry == nil {
		fmt.Printf("error: unknown command %q\n", name)
		os.Exit(1)
	}

	cmdFlags := entry.Flags(flag.NewFlagSet(entry.Name, flag.ExitOnError))
	cmdFlags.Usage = func() { usage(cmdFlags) }
	cmdFlags.Parse(arguments[1:])

	if err := entry.Run(cmdFlags.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "goose: %v\n", err)
		os.Exit(1)
	}
}

func usage(fs *flag.FlagSet) {
	fmt.Printf("Usage: goose %s [options]\n", fs.Name())
	fs.PrintDefaults()
}
