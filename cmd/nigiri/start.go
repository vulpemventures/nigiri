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
		&cli.BoolFlag{
			Name:  "signet",
			Usage: "start Bitcoin node in signet",
		},
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
	isSignet := ctx.Bool("signet")
	runWithoutFrontend := ctx.Bool("ci")
	datadir := ctx.String("datadir")
	regtestComposePath := filepath.Join(datadir, config.DefaultCompose)
	signetComposePath := filepath.Join(datadir, config.SignetCompose)

	composePath := regtestComposePath
	if isSignet {
		composePath = signetComposePath
	}

	// spin up all the services in the compose file
	bashCmd := exec.Command("docker-compose", "-f", composePath, "up", "-d", "esplora")
	if isLiquid {
		//this will only run chopsticks & chopsticks-liquid and servives they depends on
		bashCmd = exec.Command("docker-compose", "-f", composePath, "up", "-d", "esplora", "esplora-liquid")
	}

	if runWithoutFrontend {
		//this will only run chopsticks and servives it depends on
		bashCmd = exec.Command("docker-compose", "-f", composePath, "up", "-d", "chopsticks")
		if isLiquid {
			//this will only run chopsticks & chopsticks-liquid and servives they depends on
			bashCmd = exec.Command("docker-compose", "-f", composePath, "up", "-d", "chopsticks", "chopsticks-liquid")
		}
	}

	if isSignet {
		if isLiquid || runWithoutFrontend {
			return errors.New("signet can only be used with bitcoin")
		}
		bashCmd = exec.Command("docker-compose", "-f", composePath, "up", "-d", "esplora")
	}

	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		return err
	}

	if err := nigiriState.Set(map[string]string{
		"running": strconv.FormatBool(true),
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
