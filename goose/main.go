package main

import (
	"os"

	"github.com/mei-rune/goose"
)

func main() {
	goose.Run(os.Args[1:]...)
}
