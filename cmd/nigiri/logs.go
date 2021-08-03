package main

import (
	"errors"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
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

	isRunning, err := getBoolFromState("running")
	if err != nil {
		return err
	}

	if !isRunning {
		return errors.New("nigiri is not running")
	}

	if ctx.NArg() != 1 {
		return errors.New("missing service name")
	}

	serviceName := ctx.Args().First()

	isLiquid := ctx.Bool("liquid")
	composePath := getCompose(isLiquid)

	bashCmd := exec.Command("docker-compose", "-f", composePath, "logs", serviceName)
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		return err
	}

	return nil
}
