package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vulpemventures/nigiri/cli/constants"
	"github.com/vulpemventures/nigiri/cli/controller"
)

var (
	version    = "dev"
	commit     = "none"
	date       = "unknown"
	VersionCmd = &cobra.Command{
		Use:     "version",
		Short:   "Show the current version",
		RunE:    versionFunc,
		PreRunE: versionChecks,
	}
)

func versionChecks(cmd *cobra.Command, args []string) error {
	datadir, _ := cmd.Flags().GetString("datadir")

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

	return nil
}

func versionFunc(cmd *cobra.Command, address []string) error {
	fmt.Println("Version: " + version)
	fmt.Println("Commit: " + commit)
	fmt.Println("Date: " + date)
	return nil
}
