package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/urfave/cli/v2"
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
	isLiquid, err := nigiriState.GetBool("attachliquid")
	if err != nil {
		return err
	}
	datadir := ctx.String("datadir")
	composePath := getCompose(datadir, isLiquid)

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
		if err := os.RemoveAll(ctx.String("datadir")); err != nil {
			return err
		}

		if err := provisionResourcesToDatadir(ctx.String("datadir")); err != nil {
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
