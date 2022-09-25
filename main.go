package main

import (
	"os"
	"runtime"

	"github.com/mrtc0/noic/cmd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	version   = "unknown"
	gitCommit = ""
)

const (
	appName = "noic"
	usage   = `Toy low-layer container runtime

To start a new instance of a container:

	# noic run [rootfs]

	`
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = usage

	v := []string{version}

	if gitCommit != "" {
		v = append(v, "commit: "+gitCommit)
	}
	v = append(v, "go: "+runtime.Version())

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug logging",
		},
	}

	app.Commands = []cli.Command{
		cmd.InitCommand,
		cmd.CreateCommand,
		cmd.StartCommand,
		cmd.RunCommand,
		cmd.ListCommand,
		cmd.DeleteCommand,
		cmd.StateCommand,
	}

	app.Before = func(context *cli.Context) error {
		setupLogger(context)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func setupLogger(context *cli.Context) error {
	if context.GlobalBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}
	return nil
}
