package main

import (
	"bitbucket.org/liamstask/goose"
	"os"
)

func main() {
	goose.Run(os.Args[1:]...)
}
