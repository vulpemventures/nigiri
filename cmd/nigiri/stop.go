package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/nigiri/internal/config"
)

var stop = cli.Command{
	Name:   "stop",
	Usage:  "stop nigiri",
	Action: stopAction,
	Flags: []cli.Flag{
		&liquidFlag,
		&cli.BoolFlag{
			Name:  "delete",
			Usage: "clean node data directories",
			Value: false,
		},
	},
}

func stopAction(ctx *cli.Context) error {

	delete := ctx.Bool("delete")
	datadir := ctx.String("datadir")
	composePath := filepath.Join(datadir, config.DefaultCompose)

	bashCmd := exec.Command("docker-compose", "-f", composePath, "stop")
	if delete {
		bashCmd = exec.Command("docker-compose", "-f", composePath, "down", "--volumes")
	}
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		return err
	}

	if delete {
		fmt.Println("Removing data from volumes...")

		datadir := ctx.String("datadir")
		if err := os.RemoveAll(datadir); err != nil {
			return err
		}

		if err := provisionResourcesToDatadir(datadir); err != nil {
			return err
		}

		fmt.Println("Nigiri has been cleaned up successfully.")
	} else {
		if err := nigiriState.Set(map[string]string{
			"running": strconv.FormatBool(false),
		}); err != nil {
			return err
		}
	}

	return nil
}
