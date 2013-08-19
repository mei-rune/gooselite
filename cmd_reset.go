package goose

import (
	"fmt"
	"log"
)

var resetCmd = &Command{
	Name:    "reset",
	Usage:   "",
	Summary: "Roll back the version to 0 and Migrate the DB to the most recent version available",
	Help:    `reset extended help here...`,
}

func resetRun(cmd *Command, args ...string) {
	conf, err := NewDBConf()
	if err != nil {
		log.Fatal(err)
	}

	current := GetDBVersion(conf)
	if current != 0 {
		RunMigrations(conf, conf.MigrationsDir, 0)
	}

	target := MostRecentVersionAvailable(conf.MigrationsDir)
	RunMigrations(conf, conf.MigrationsDir, target)
}

func init() {
	resetCmd.Run = resetRun
	Commands = append(Commands, resetCmd)
}
