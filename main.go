package main

import (
	"os"
	"runtime"

	"github.com/urfave/cli"
	"github.com/mrtc0/noic/cmd"
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
		cmd.RunCommand,
	}

	app.Before = func(context *cli.Context) error {
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
