package cmd

import (
	"os"
	"os/exec"

	"github.com/vulpemventures/nigiri/cli/constants"
	"github.com/vulpemventures/nigiri/cli/controller"

	"github.com/spf13/cobra"
)

var LogsCmd = &cobra.Command{
	Use:     "logs [command]",
	Short:   "Check service logs",
	RunE:    logs,
	PreRunE: logsChecks,
}

var NodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Check logs for node service",
	RunE: func(cmd *cobra.Command, args []string) error {
		logs(cmd, []string{"node"})
		return nil
	},
}
var ElectrsCmd = &cobra.Command{
	Use:   "electrs",
	Short: "Check logs for electrs",
	RunE: func(cmd *cobra.Command, args []string) error {
		logs(cmd, []string{"electrs"})
		return nil
	},
}
var ChopsticksCmd = &cobra.Command{
	Use:   "chopsticks",
	Short: "Check logs for chopsticks",
	RunE: func(cmd *cobra.Command, args []string) error {
		logs(cmd, []string{"chopsticks"})
		return nil
	},
}
var EsploraCmd = &cobra.Command{
	Use:   "esplora",
	Short: "Check logs for esplora",
	RunE: func(cmd *cobra.Command, args []string) error {
		logs(cmd, []string{"esplora"})
		return nil
	},
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
