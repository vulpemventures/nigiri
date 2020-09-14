package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/vulpemventures/nigiri/cli/constants"
	"github.com/vulpemventures/nigiri/cli/controller"
)

var RpcCmd = &cobra.Command{
	Use:     "rpc <command>",
	Short:   "Wrapper for accessing the bitcoin-cli and elements-cli",
	RunE:    rpc,
	PreRunE: rpcChecks,
}

func rpcChecks(cmd *cobra.Command, args []string) error {
	datadir, _ := cmd.Flags().GetString("datadir")
	isLiquidService, _ := cmd.Flags().GetBool("liquid")

	ctl, err := controller.NewController()
	if err != nil {
		return err
	}

	if err := ctl.ParseDatadir(datadir); err != nil {
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

func rpc(cmd *cobra.Command, args []string) error {
	datadir, _ := cmd.Flags().GetString("datadir")
	isLiquidService, _ := cmd.Flags().GetBool("liquid")

	ctl, err := controller.NewController()
	if err != nil {
		return err
	}

	envPath := ctl.GetResourcePath(datadir, "env")
	env := ctl.LoadComposeEnvironment(envPath)

	rpcArgs := []string{"exec", "bitcoin", "bitcoin-cli", "-datadir=config"}
	if isLiquidService {
		rpcArgs = []string{"exec", "liquid", "elements-cli", "-datadir=config"}
	}
	cmdArgs := append(rpcArgs, args...)
	bashCmd := exec.Command("docker", cmdArgs...)
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr
	bashCmd.Env = env

	if err := bashCmd.Run(); err != nil {
		return err
	}

	return nil
}
