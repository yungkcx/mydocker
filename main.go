package main

import (
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"

	app.Commands = []cli.Command{
		initCommand,
		runCommand,
		imagesCommand,
		imageCommand,
		psCommand,
		rmiCommand,
	}

	app.Before = func(context *cli.Context) error {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err.Error())
	}
}
