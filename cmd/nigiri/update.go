package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/nigiri/internal/config"
)

var update = cli.Command{
	Name:   "update",
	Usage:  "check for updates and pull new docker images",
	Action: updateAction,
}

func updateAction(ctx *cli.Context) error {
	datadir := ctx.String("datadir")
	composePath := filepath.Join(datadir, config.DefaultCompose)

	bashCmd := exec.Command("docker-compose", "-f", composePath, "pull")
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		return err
	}

	return nil
}
