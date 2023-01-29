package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/nigiri/internal/config"
)

var logs = cli.Command{
	Name:   "logs",
	Usage:  "check Service logs",
	Action: logsAction,
	Flags: []cli.Flag{
		&liquidFlag,
	},
}

func logsAction(ctx *cli.Context) error {

	if isRunning, _ := nigiriState.GetBool("running"); !isRunning {
		return errors.New("nigiri is not running")
	}

	if ctx.NArg() != 1 {
		return errors.New("missing service name")
	}

	serviceName := ctx.Args().First()

	datadir := ctx.String("datadir")
	composePath := filepath.Join(datadir, config.DefaultCompose)

	bashCmd := runDockerCompose(composePath, "logs", serviceName)
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		return err
	}

	return nil
}
