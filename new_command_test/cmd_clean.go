package main

import (
	"bitbucket.org/liamstask/goose"
	"fmt"
	"log"
)

var cleanCmd = &goose.Command{
	Name:    "clean",
	Usage:   "",
	Summary: "Roll back the version by 1",
	Help:    `clean extended help here...`,
}

func cleanRun(cmd *goose.Command, args ...string) {
	conf, err := goose.NewDBConf()
	if err != nil {
		log.Fatal(err)
	}

	current := goose.GetDBVersion(conf)
	if current == 0 {
		fmt.Println("db is empty, can't clean.")
		return
	}

	goose.RunMigrations(conf, conf.MigrationsDir, 0)
}

func init() {
	cleanCmd.Run = cleanRun
	goose.Commands = append(goose.Commands, cleanCmd)
}
