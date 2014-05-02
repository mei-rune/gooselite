package goose

import (
	"log"
)

var resetCmd = &Command{
	Name:    "reset",
	Usage:   "",
	Summary: "Roll back the version to 0 and Migrate the DB to the most recent version available",
	Help:    `reset extended help here...`,
	Run:     resetRun,
}

func resetRun(cmd *Command, args ...string) {
	conf, err := dbConfFromFlags()
	if err != nil {
		log.Fatal(err)
	}

	current, err := GetDBVersion(conf)
	if err != nil {
		log.Fatal(err)
	}

	if current != 0 {
		RunMigrations(conf, conf.MigrationsDir, 0)
	}

	target, err := GetMostRecentDBVersion(conf.MigrationsDir)
	if err != nil {
		log.Fatal(err)
	}
	err = RunMigrations(conf, conf.MigrationsDir, target)
	if err != nil {
		log.Fatal(err)
	}
}
