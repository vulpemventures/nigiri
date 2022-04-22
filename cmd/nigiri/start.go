package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

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
		&cli.BoolFlag{
			Name:  "ci",
			Usage: "runs in headless mode without esplora for continuous integration environments",
			Value: false,
		},
	},
}

func startAction(ctx *cli.Context) error {

	if isRunning, _ := nigiriState.GetBool("running"); isRunning {
		return errors.New("nigiri is already running, please stop it first")
	}

	isLiquid := ctx.Bool("liquid")
	isLN := ctx.Bool("ln")
	isCI := ctx.Bool("ci")
	datadir := ctx.String("datadir")
	composePath := filepath.Join(datadir, config.DefaultCompose)

	// spin up all the services in the compose file
	servicesToRun := []string{"esplora"}
	if isLiquid {
		//this will only run chopsticks & chopsticks-liquid and servives they depends on
		servicesToRun = append(servicesToRun, "esplora-liquid")
	}

	if isLN {
		// LND
		servicesToRun = append(servicesToRun, "lnd")
		// Core Lightning Network
		servicesToRun = append(servicesToRun, "lightningd")

		servicesToRun = append(servicesToRun, "sensei")
	}

	if isCI {
		//this will only run chopsticks and servives it depends on
		servicesToRun = []string{"chopsticks"}
		if isLiquid {
			//this will only run chopsticks & chopsticks-liquid and servives they depends on
			servicesToRun = append(servicesToRun, "chopsticks-liquid")
		}
		// add also LN services if needed
		if isLN {
			// LND
			servicesToRun = append(servicesToRun, "lnd")
			// Core Lightning Network
			servicesToRun = append(servicesToRun, "lightningd")
		}
	}

	args := []string{"-f", composePath, "up", "-d"}
	args = append(args, servicesToRun...)

	bashCmd := exec.Command("docker-compose", args...)
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		return err
	}

	if err := nigiriState.Set(map[string]string{
		"running": strconv.FormatBool(true),
		"ci":      strconv.FormatBool(isCI),
		"liquid":  strconv.FormatBool(isLiquid),
		"ln":      strconv.FormatBool(isLN),
	}); err != nil {
		return err
	}

	services, err := docker.GetServices(composePath)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("ENDPOINTS")

	for _, nameAndEndpoint := range services {
		name := nameAndEndpoint[0]
		endpoint := nameAndEndpoint[1]

		if !isLiquid && strings.Contains(name, "liquid") {
			continue
		}

		fmt.Println(name + " " + endpoint)
	}

	return nil
}
