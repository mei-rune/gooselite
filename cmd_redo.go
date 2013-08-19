package goose

import "log"

var redoCmd = &Command{
	Name:    "redo",
	Usage:   "",
	Summary: "Re-run the latest migration",
	Help:    `redo extended help here...`,
}

func redoRun(cmd *Command, args ...string) {
	conf, err := NewDBConf()
	if err != nil {
		log.Fatal(err)
	}

	target := GetDBVersion(conf)
	_, earliest := GetPreviousVersion(conf.MigrationsDir, target)

	downRun(cmd, args...)
	if target == 0 {
		log.Printf("Updating from %s to %s\n", target, earliest)
		target = earliest
	}
	RunMigrations(conf, conf.MigrationsDir, target)
}

func init() {
	redoCmd.Run = redoRun
}
