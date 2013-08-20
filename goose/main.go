package main

import (
	"bitbucket.org/runner_mei/goose"
	"os"
)

func main() {
	goose.Run(os.Args[1:]...)
}
