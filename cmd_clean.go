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
}

func cleanRun(cmd *Command, args ...string) {
	conf, err := NewDBConf()
	if err != nil {
		log.Fatal(err)
	}

	current := GetDBVersion(conf)
	if current == 0 {
		fmt.Println("db is empty, can't clean.")
		return
	}

	RunMigrations(conf, conf.MigrationsDir, 0)
}

func init() {
	cleanCmd.Run = cleanRun
	Commands = append(Commands, cleanCmd)
}
