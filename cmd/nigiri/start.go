package main

import (
	"errors"
	"os"
	"os/exec"
	"strconv"

	"github.com/urfave/cli/v2"
)

var start = cli.Command{
	Name:   "start",
	Usage:  "start nigiri",
	Action: startAction,
	Flags: []cli.Flag{
		&liquidFlag,
	},
}

func startAction(ctx *cli.Context) error {

	if isRunning, _ := nigiriState.GetBool("running"); isRunning {
		return errors.New("nigiri is already running, please stop it first")
	}

	isLiquid := ctx.Bool("liquid")
	composePath := getCompose(isLiquid)

	bashCmd := exec.Command("docker-compose", "-f", composePath, "up", "-d")
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		return err
	}

	if err := nigiriState.Set(map[string]string{
		"attachliquid": strconv.FormatBool(isLiquid),
		"running":      strconv.FormatBool(true),
	}); err != nil {
		return err
	}

	return nil
}
