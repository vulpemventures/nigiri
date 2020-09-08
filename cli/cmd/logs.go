package cmd

import (
	"errors"
	"os"
	"os/exec"

	"github.com/vulpemventures/nigiri/cli/constants"
	"github.com/vulpemventures/nigiri/cli/controller"

	"github.com/spf13/cobra"
)

var logsDescription = "Check Service logs. Requires one Service: " + servicesList()

var LogsCmd = &cobra.Command{
	Args: func(cmd *cobra.Command, args []string) error {
		_, found := controller.Services[args[0]]
		if !found {
			return errors.New(logsDescription)
		}
		return nil
	},
	Use:     "logs <service>",
	Short:   logsDescription,
	Long:    logsDescription,
	RunE:    logs,
	PreRunE: logsChecks,
}

func servicesList() string {
	var servicesString string
	for key, _ := range controller.Services {
		servicesString += key
		servicesString += " "
	}
	return servicesString[:len(servicesString)-1]
}

func logsChecks(cmd *cobra.Command, args []string) error {
	datadir, _ := cmd.Flags().GetString("datadir")
	isLiquidService, _ := cmd.Flags().GetBool("liquid")

	ctl, err := controller.NewController()
	if err != nil {
		return err
	}

	if err := ctl.ParseDatadir(datadir); err != nil {
		return err
	}
	if len(args) != 1 {
		return constants.ErrInvalidArgs
	}

	service := args[0]
	if err := ctl.ParseServiceName(service); err != nil {
		return err
	}

	if isRunning, err := ctl.IsNigiriRunning(); err != nil {
		return err
	} else if !isRunning {
		return constants.ErrNigiriNotRunning
	}

	if err := ctl.ReadConfigFile(datadir); err != nil {
		return err
	}

	if isLiquidService && isLiquidService != ctl.GetConfigBoolField(constants.AttachLiquid) {
		return constants.ErrNigiriLiquidNotEnabled
	}

	return nil
}

func logs(cmd *cobra.Command, args []string) error {
	service := args[0]
	datadir, _ := cmd.Flags().GetString("datadir")
	isLiquidService, _ := cmd.Flags().GetBool("liquid")

	ctl, err := controller.NewController()
	if err != nil {
		return err
	}

	serviceName := ctl.GetServiceName(service, isLiquidService)
	composePath := ctl.GetResourcePath(datadir, "compose")
	envPath := ctl.GetResourcePath(datadir, "env")
	env := ctl.LoadComposeEnvironment(envPath)

	bashCmd := exec.Command("docker-compose", "-f", composePath, "logs", serviceName)
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr
	bashCmd.Env = env

	if err := bashCmd.Run(); err != nil {
		return err
	}

	return nil
}
