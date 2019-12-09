package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/vulpemventures/nigiri/cli/constants"
	"github.com/vulpemventures/nigiri/cli/controller"
)

var StopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stop and/or delete nigiri",
	RunE:    stop,
	PreRunE: stopChecks,
}

func stopChecks(cmd *cobra.Command, args []string) error {
	datadir, _ := cmd.Flags().GetString("datadir")
	delete, _ := cmd.Flags().GetBool("delete")

	ctl, err := controller.NewController()
	if err != nil {
		return err
	}

	if err := ctl.ParseDatadir(datadir); err != nil {
		return err
	}

	if _, err := os.Stat(datadir); os.IsNotExist(err) {
		return constants.ErrDatadirNotExisting
	}

	if isRunning, err := ctl.IsNigiriRunning(); err != nil {
		return err
	} else if !isRunning {
		if delete {
			if isStopped, err := ctl.IsNigiriStopped(); err != nil {
				return err
			} else if !isStopped {
				return constants.ErrNigiriNotExisting
			}
		} else {
			return constants.ErrNigiriNotRunning
		}
	}

	if err := ctl.ReadConfigFile(datadir); err != nil {
		return err
	}
	return nil
}

func stop(cmd *cobra.Command, args []string) error {
	delete, _ := cmd.Flags().GetBool("delete")
	datadir, _ := cmd.Flags().GetString("datadir")

	ctl, err := controller.NewController()
	if err != nil {
		return err
	}

	composePath := ctl.GetResourcePath(datadir, "compose")
	configPath := ctl.GetResourcePath(datadir, "config")
	envPath := ctl.GetResourcePath(datadir, "env")
	env := ctl.LoadComposeEnvironment(envPath)

	bashCmd := exec.Command("docker-compose", "-f", composePath, "stop")
	if delete {
		bashCmd = exec.Command("docker-compose", "-f", composePath, "down")
	}
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr
	bashCmd.Env = env

	if err := bashCmd.Run(); err != nil {
		return err
	}

	if delete {
		fmt.Println("Removing data from volumes...")
		if err := ctl.CleanResourceVolumes(datadir); err != nil {
			return err
		}

		fmt.Println("Removing configuration file...")
		if err := os.Remove(configPath); err != nil {
			return err
		}

		fmt.Println("Removing environmet file...")
		if err := os.Remove(envPath); err != nil {
			return err
		}

		fmt.Println("Nigiri has been cleaned up successfully.")
	}

	return nil
}
