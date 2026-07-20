package main

import (
	"fmt"
	"os"

	"github.com/mei-rune/goose"
)

func main() {
	if err := goose.Run(os.Args[1:]...); err != nil {
		fmt.Fprintf(os.Stderr, "goose: %v\n", err)
		os.Exit(1)
	}
}
