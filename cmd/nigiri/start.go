package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/nigiri/internal/config"
)

var start = cli.Command{
	Name:   "start",
	Usage:  "start nigiri",
	Action: startAction,
	Flags: []cli.Flag{
		&liquidFlag,
		&lnFlag,
		&arkFlag,
	},
}

func startAction(ctx *cli.Context) error {

	if isRunning, _ := nigiriState.GetBool("running"); isRunning {
		return errors.New("nigiri is already running, please stop it first")
	}

	datadir := ctx.String("datadir")
	composePath := filepath.Join(datadir, config.DefaultCompose)

	// Build the docker-compose command with appropriate services
	services := []string{"bitcoin", "electrs", "chopsticks", "esplora"}

	if ctx.Bool("liquid") {
		services = append(services, "liquid", "electrs-liquid", "chopsticks-liquid", "esplora-liquid")
	}

	if ctx.Bool("ln") {
		services = append(services, "lnd", "tap", "cln")
	}

	if ctx.Bool("ark") {
		services = append(services, "ark")
	}

	// Start the services
	bashCmd := runDockerCompose(composePath, append([]string{"up", "-d"}, services...)...)
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		return err
	}

	// Update state
	fmt.Printf("🍣 nigiri configuration located at %s\n", nigiriState.FilePath())
	if err := nigiriState.Set(map[string]string{
		"running": strconv.FormatBool(true),
		"liquid":  strconv.FormatBool(ctx.Bool("liquid")),
		"ln":      strconv.FormatBool(ctx.Bool("ln")),
		"ark":     strconv.FormatBool(ctx.Bool("ark")),
	}); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	return nil
}
