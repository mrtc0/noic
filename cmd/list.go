package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var ListCommand = cli.Command{
	Name:  "list",
	Usage: "list containers",
	Action: func(context *cli.Context) error {
		stateRootDirectory := context.GlobalString("root")

		files, err := os.ReadDir(stateRootDirectory)
		if err != nil {
			logrus.Errorf("Read dir %s error %v", stateRootDirectory, err)
			return err
		}

		for _, file := range files {
			fmt.Println(file.Name())
		}

		return nil
	},
}
