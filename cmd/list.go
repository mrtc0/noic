package cmd

import (
	"fmt"
	"os"

	"github.com/mrtc0/noic/pkg/container"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var ListCommand = cli.Command{
	Name:  "list",
	Usage: "list containers",
	Action: func(context *cli.Context) error {
		files, err := os.ReadDir(container.StateDir)
		if err != nil {
			logrus.Errorf("Read dir %s error %v", container.StateDir, err)
			return err
		}

		for _, file := range files {
			fmt.Println(file.Name())
		}

		return nil
	},
}
