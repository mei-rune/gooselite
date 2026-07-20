package goose

import (
	"flag"
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
