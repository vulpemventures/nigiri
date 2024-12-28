package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/nigiri/internal/config"
	"github.com/vulpemventures/nigiri/internal/docker"
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

	// Get endpoints from docker-compose
	client := docker.NewDefaultClient()
	endpoints, err := client.GetEndpoints(composePath)
	if err != nil {
		return fmt.Errorf("failed to get endpoints: %w", err)
	}

	// Filter endpoints based on enabled services
	filteredEndpoints := make(map[string]string)
	for name, endpoint := range endpoints {
		if !ctx.Bool("liquid") && strings.Contains(name, "liquid") {
			continue
		}
		if !ctx.Bool("ln") && (strings.Contains(name, "lnd") || strings.Contains(name, "cln") || strings.Contains(name, "tap")) {
			continue
		}
		if !ctx.Bool("ark") && strings.Contains(name, "ark") {
			continue
		}
		filteredEndpoints[name] = endpoint
	}

	// Display endpoints
	fmt.Println("\n🍜 ENDPOINTS")
	for name, endpoint := range filteredEndpoints {
		fmt.Printf("%s %s: %s\n", 
			aurora.Green("✓"),
			aurora.Blue(name),
			endpoint,
		)
	}

	return nil
}
