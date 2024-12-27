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

	bashCmd := runDockerCompose(composePath, "stop")
	if delete {
		cleanupCmd := runDockerCompose(composePath, "run", "-T", "--rm", "--entrypoint", "sh", "bitcoin", "-c", "chown -R $(id -u):$(id -g) /data/.bitcoin")
		cleanupCmd.Stdout = os.Stdout
		cleanupCmd.Stderr = os.Stderr
		if err := cleanupCmd.Run(); err != nil {
			fmt.Printf("Warning: cleanup container failed: %v\n", err)
		}

		bashCmd = runDockerCompose(composePath, "down", "--volumes")
	}

	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		return err
	}

	if delete {
		fmt.Println("Removing data from volumes...")

		if err := os.RemoveAll(datadir); err != nil {
			fmt.Printf("Warning: could not remove data directory: %v\n", err)
			sudoCmd := exec.Command("sudo", "rm", "-rf", datadir)
			if err := sudoCmd.Run(); err != nil {
				return fmt.Errorf("failed to remove data directory even with sudo: %w", err)
			}
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
