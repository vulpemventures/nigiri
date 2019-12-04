package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vulpemventures/nigiri/cli/constants"
	"github.com/vulpemventures/nigiri/cli/controller"
)

var StartCmd = &cobra.Command{
	Use:     "start",
	Short:   "Build and start Nigiri",
	RunE:    start,
	PreRunE: startChecks,
}

func startChecks(cmd *cobra.Command, args []string) error {
	network, _ := cmd.Flags().GetString("network")
	datadir, _ := cmd.Flags().GetString("datadir")
	env, _ := cmd.Flags().GetString("env")

	ctl, err := controller.NewController()
	if err != nil {
		return err
	}

	if err := ctl.ParseNetwork(network); err != nil {
		return err
	}

	if err := ctl.ParseDatadir(datadir); err != nil {
		return err
	}
	composeEnv, err := ctl.ParseEnv(env)
	if err != nil {
		return err
	}

	// if nigiri is already running return error
	if isRunning, err := ctl.IsNigiriRunning(); err != nil {
		return err
	} else if isRunning {
		return constants.ErrNigiriAlreadyRunning
	}

	// scratch datadir if not exists
	if err := os.MkdirAll(datadir, 0755); err != nil {
		return err
	}

	// if datadir is set we must copy the resources directory from ~/.nigiri
	// to the new one
	if datadir != ctl.GetDefaultDatadir() {
		if err := ctl.NewDatadirFromDefault(datadir); err != nil {
			return err
		}
	}

	// if nigiri not exists, we need to write the configuration file and then
	// read from it to get viper updated, otherwise we just read from it.
	if isStopped, err := ctl.IsNigiriStopped(); err != nil {
		return err
	} else if isStopped {
		if err := ctl.ReadConfigFile(datadir); err != nil {
			return err
		}
	} else {
		filedir := ctl.GetResourcePath(datadir, "config")
		if err := ctl.WriteConfigFile(filedir); err != nil {
			return err
		}
		// .env must be in the directory where docker-compose is run from, not where YAML files are placed
		// https://docs.docker.com/compose/env-file/
		filedir = ctl.GetResourcePath(datadir, "env")
		if err := ctl.WriteComposeEnvironment(filedir, composeEnv); err != nil {
			return err
		}
	}

	return nil
}

func start(cmd *cobra.Command, args []string) error {
	ctl, err := controller.NewController()
	if err != nil {
		return err
	}

	datadir, _ := cmd.Flags().GetString("datadir")
	liquidEnabled := ctl.GetConfigBoolField(constants.AttachLiquid)

	envPath := ctl.GetResourcePath(datadir, "env")
	composePath := ctl.GetResourcePath(datadir, "compose")

	bashCmd := exec.Command("docker-compose", "-f", composePath, "up", "-d")
	if isStopped, err := ctl.IsNigiriStopped(); err != nil {
		return err
	} else if isStopped {
		bashCmd = exec.Command("docker-compose", "-f", composePath, "start")
	}
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr
	bashCmd.Env = ctl.LoadComposeEnvironment(envPath)

	if err := bashCmd.Run(); err != nil {
		return err
	}

	path := ctl.GetResourcePath(datadir, "env")
	env, err := ctl.ReadComposeEnvironment(path)
	if err != nil {
		return err
	}

	prettyPrintServices := func(chain string, services map[string]int) {
		fmt.Printf("%s services:\n", strings.Title(chain))
		for name, port := range services {
			formatName := fmt.Sprintf("%s:", name)
			fmt.Printf("   %-14s localhost:%d\n", formatName, port)
		}
	}

	for chain, services := range env["ports"].(map[string]map[string]int) {
		if chain == "bitcoin" {
			prettyPrintServices(chain, services)
		} else if liquidEnabled {
			prettyPrintServices(chain, services)
		}
	}

	return nil
}
