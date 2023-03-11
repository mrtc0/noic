package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mrtc0/noic/cmd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	version   = "unknown"
	gitCommit = ""

	stateRootDirectory = "/run/noic"
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
		cli.StringFlag{
			Name:  "root",
			Value: stateRootDirectory,
			Usage: "root directory for storage of container state (this should be located in tmpfs)",
		},
		cli.StringFlag{
			Name:  "log",
			Value: "",
			Usage: "set the log file to write noic logs to (default is '/dev/stderr')",
		},
		cli.StringFlag{
			Name:  "log-format",
			Value: "text",
			Usage: "set the log format ('text' (default), or 'json')",
		},
		cli.StringFlag{
			Name:   "criu",
			Usage:  "Not Supported",
			Hidden: true,
		},
		cli.BoolFlag{
			Name:  "systemd-cgroup",
			Usage: "enable systemd cgroup support, expects cgroupsPath to be of form \"slice:prefix:name\" for e.g. \"system.slice:runc:434234\"",
		},
		cli.StringFlag{
			Name:  "rootless",
			Value: "auto",
			Usage: "not supported",
		},
	}

	app.Commands = []cli.Command{
		cmd.InitCommand,
		cmd.CreateCommand,
		cmd.StartCommand,
		cmd.RunCommand,
		cmd.ListCommand,
		cmd.DeleteCommand,
		cmd.KillCommand,
		cmd.StateCommand,
		cmd.ExecCommand,
	}

	app.Before = func(context *cli.Context) error {
		if err := setupLogger(context); err != nil {
			return err
		}

		if !context.IsSet("root") {
			if err := os.MkdirAll(stateRootDirectory, 0o700); err != nil {
				return err
			}

			if err := os.Chmod(stateRootDirectory, os.FileMode(0o700)|os.ModeSticky); err != nil {
				return err
			}
		}

		if err := stateRootDirectoryToAbsolutePath(context); err != nil {
			return err
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func stateRootDirectoryToAbsolutePath(context *cli.Context) error {
	p := context.GlobalString("root")
	if p == "" {
		return nil
	}

	p, err := filepath.Abs(p)
	if err != nil {
		return err
	}

	return context.GlobalSet("root", p)
}

func setupLogger(context *cli.Context) error {
	if context.GlobalBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	switch logFormat := context.GlobalString("log-format"); logFormat {
	case "":
	case "text":
	case "json":
		logrus.SetFormatter(new(logrus.JSONFormatter))
	default:
		return fmt.Errorf("invalid log-format: %s", logFormat)
	}

	if logFile := context.GlobalString("log"); logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0o644)
		if err != nil {
			return err
		}

		logrus.SetOutput(f)
	}

	return nil
}
