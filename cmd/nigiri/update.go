package main

import (
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var update = cli.Command{
	Name:   "update",
	Usage:  "check for updates and pull new docker images",
	Action: updateAction,
}

func updateAction(ctx *cli.Context) error {
	composePath := getCompose(true)

	bashCmd := exec.Command("docker-compose", "-f", composePath, "pull")
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		return err
	}

	return nil
}
