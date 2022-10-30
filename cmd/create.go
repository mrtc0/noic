package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mrtc0/noic/pkg/container"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var CreateCommand = cli.Command{
	Name:  "create",
	Usage: "create a container",
	ArgsUsage: `<container-id>

Where "<container-id>" is your name for instance of the container.
	`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "bundle, b",
			Value: ".",
			Usage: `path to the bundle directory, defaults to the current directory`,
		},
		cli.StringFlag{
			Name:  "pid-file",
			Value: "",
			Usage: "specify the file to write the process id to",
		},
	},
	Action: func(context *cli.Context) error {
		containerID := context.Args().First()
		stateRootDirectory := context.GlobalString("root")

		pidFile := context.String("pid-file")
		if pidFile != "" {
			pidFile, err := filepath.Abs(pidFile)
			if err != nil {
				return err
			}

			context.Set("pid-file", pidFile)
		}

		if container.Exists(stateRootDirectory, containerID) {
			return fmt.Errorf("container %s is exists", containerID)
		}

		bundlePath := context.String("bundle")

		factory := &container.ContainerFactory{
			ContainerID:        containerID,
			StateRootDirectory: stateRootDirectory,
			BundlePath:         bundlePath,
		}

		c, err := factory.Create()
		if err != nil {
			return err
		}

		logrus.Info("Run parent process")
		if err := c.Run(); err != nil {
			return fmt.Errorf("run failed: %v", err)
		}
		logrus.Info("Save state")
		if err = c.SaveState(); err != nil {
			return fmt.Errorf("failed save state: %v", err)
		}

		logrus.Info("Create PID file")
		if context.IsSet("pid-file") {
			if err = c.CreatePIDFile(context.String("pid-file")); err != nil {
				return err
			}
		}

		logrus.Info("create is done")
		return nil
	},
}
