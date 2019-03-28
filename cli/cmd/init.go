package cmd

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vulpemventures/nigiri/cli/config"
)

var InitCmd = &cobra.Command{
	Use:    "init",
	Short:  "Initialize configuration file",
	Run:    run,
	PreRun: runChecks,
}

func run(cmd *cobra.Command, args []string) {
	viper := config.Viper()

	if err := viper.WriteConfig(); err != nil {
		log.WithError(err).Fatal("An error occured while writing config file")
	}
}

func runChecks(cmd *cobra.Command, args []string) {
	network := cmd.Flags().Lookup("network").Value.String()

	// check valid network
	if !isNetworkOk(network) {
		log.WithField("network_flag", network).Fatal("Invalid network")
	}

	// scratch ~/.nigiri/ if not exists
	if err := os.MkdirAll(config.GetPath(), 0755); err != nil {
		log.WithError(err).Fatal("An error occured while scratching config dir")
	}

	// check if config file already exists
	if _, err := os.Stat(config.GetFullPath()); !os.IsNotExist(err) {
		log.Fatal("File already exists, please delete it first")
	}
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

func fileExists(path string) bool {
	_, err := os.Stat(filepath.Join(path, config.Filename))
	return !os.IsNotExist(err)
}
