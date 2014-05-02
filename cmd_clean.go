package goose

import (
	"fmt"
	"log"
)

var cleanCmd = &Command{
	Name:    "clean",
	Usage:   "",
	Summary: "Roll back the version by 1",
	Help:    `clean extended help here...`,
	Run:     cleanRun,
}

func cleanRun(cmd *Command, args ...string) {
	conf, err := dbConfFromFlags()
	if err != nil {
		log.Fatal(err)
	}

	current, err := GetDBVersion(conf)
	if err != nil {
		log.Fatal(err)
	}

	if current == 0 {
		fmt.Println("db is empty, can't clean.")
		return
	}

	err = RunMigrations(conf, conf.MigrationsDir, 0)
	if err != nil {
		log.Fatal(err)
	}
}
