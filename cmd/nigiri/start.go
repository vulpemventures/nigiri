package main

import (
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var start = cli.Command{
	Name:   "start",
	Usage:  "start nigiri box",
	Action: startAction,
	Flags: []cli.Flag{
		&liquidFlag,
	},
}

func startAction(ctx *cli.Context) error {

	isLiquid := ctx.Bool("liquid")
	composePath := getCompose(isLiquid)

	bashCmd := exec.Command("docker-compose", "-f", composePath, "up", "-d")
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		return err
	}

	return nil
}
