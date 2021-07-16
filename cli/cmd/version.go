package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version    = "dev"
	commit     = "none"
	date       = "unknown"
	VersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the current version",
		RunE:  versionFunc,
	}
)

func versionFunc(cmd *cobra.Command, address []string) error {
	fmt.Println("Version: " + version)
	fmt.Println("Commit: " + commit)
	fmt.Println("Date: " + date)
	return nil
}
