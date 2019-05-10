package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vulpemventures/nigiri/cli/config"
)

const listAll = true

var StartCmd = &cobra.Command{
	Use:     "start",
	Short:   "Build and start Nigiri",
	RunE:    start,
	PreRunE: startChecks,
}

func startChecks(cmd *cobra.Command, args []string) error {
	network, _ := cmd.Flags().GetString("network")
	datadir, _ := cmd.Flags().GetString("datadir")

	// check flags
	if !isNetworkOk(network) {
		return fmt.Errorf("Invalid network: %s", network)
	}

	if !isDatadirOk(datadir) {
		return fmt.Errorf("Invalid datadir, it must be an absolute path: %s", datadir)
	}

	// scratch datadir if not exists
	if err := os.MkdirAll(datadir, 0755); err != nil {
		return err
	}

	// if datadir is set we must copy the resources directory from ~/.nigiri
	// to the new one
	if datadir != getDefaultDir() {
		if err := copyResources(datadir); err != nil {
			return err
		}
	}

	// if nigiri is already running return error
	isRunning, err := nigiriIsRunning()
	if err != nil {
		return err
	}
	if isRunning {
		return fmt.Errorf("Nigiri is already running, please stop it first")
	}

	// if nigiri not exists, we need to write the configuration file and then
	// read from it to get viper updated, otherwise we just read from it.
	exists, err := nigiriExistsAndNotRunning()
	if err != nil {
		return err
	}
	if !exists {
		filedir := filepath.Join(datadir, "nigiri.config.json")
		if err := config.WriteConfig(filedir); err != nil {
			return err
		}
	}
	if err := config.ReadFromFile(datadir); err != nil {
		return err
	}

	return nil
}

func start(cmd *cobra.Command, args []string) error {
	bashCmd, err := getStartBashCmd()
	if err != nil {
		return err
	}

	err = bashCmd.Run()
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"node":         "localhost:19001",
		"electrm_RPC":  "localhost:60401",
		"electrum_API": "localhost:3002",
		"esplora":      "localhost:5000",
		"chopsticks":   "localhost:3000",
	}).Info("Bitcoin services:")

	viper := config.Viper()
	if viper.GetBool(config.AttachLiquid) {
		log.WithFields(log.Fields{
			"node":         "localhost:18884",
			"electrum_RPC": "localhost:60411",
			"electrum_API": "localhost:3022",
			"esplora":      "localhost:5001",
			"chopsticks":   "localhost:3001",
		}).Info("Liquid services:")
	}

	return nil
}

var images = map[string]bool{
	"vulpemventures/bitcoin:latest":           true,
	"vulpemventures/liquid:latest":            true,
	"vulpemventures/electrs:latest":           true,
	"vulpemventures/electrs-liquid:latest":    true,
	"vulpemventures/esplora:latest":           true,
	"vulpemventures/esplora-liquid:latest":    true,
	"vulpemventures/nigiri-chopsticks:latest": true,
}

func copyResources(datadir string) error {
	defaultDatadir := getDefaultDir()
	cmd := exec.Command("cp", "-R", filepath.Join(defaultDatadir, "resources"), datadir)
	return cmd.Run()

}

func nigiriExists(listAll bool) (bool, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return false, err
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: listAll})
	if err != nil {
		return false, err
	}

	for _, container := range containers {
		if images[container.Image] {
			return true, nil
		}
	}

	return false, nil
}

func isNetworkOk(network string) bool {
	var ok bool
	for _, n := range []string{"regtest"} {
		if network == n {
			ok = true
		}
	}

	return ok
}

func isDatadirOk(datadir string) bool {
	return filepath.IsAbs(datadir)
}

func getComposePath() string {
	viper := config.Viper()
	datadir := viper.GetString("datadir")
	network := viper.GetString("network")
	attachLiquid := viper.GetBool("attachLiquid")
	if attachLiquid {
		network += "-liquid"
	}

	return filepath.Join(datadir, "resources", fmt.Sprintf("docker-compose-%s.yml", network))
}

func nigiriIsRunning() (bool, error) {
	listOnlyRunningContainers := !listAll
	return nigiriExists(listOnlyRunningContainers)
}

func nigiriExistsAndNotRunning() (bool, error) {
	return nigiriExists(listAll)
}

func getStartBashCmd() (*exec.Cmd, error) {
	composePath := getComposePath()
	bashCmd := exec.Command("docker-compose", "-f", composePath, "up", "-d")

	isStopped, err := nigiriExistsAndNotRunning()
	if err != nil {
		return nil, err
	}
	if isStopped {
		bashCmd = exec.Command("docker-compose", "-f", composePath, "start")
	}
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	return bashCmd, nil
}
