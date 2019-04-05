package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vulpemventures/nigiri/cli/builder"
	"github.com/vulpemventures/nigiri/cli/config"
	"github.com/vulpemventures/nigiri/cli/helpers"
)

var composeBuilder = map[string]func(path string) builder.ComposeBuilder{
	"regtest": helpers.NewRegtestBuilder,
}

var CreateCmd = &cobra.Command{
	Use:    "create",
	Short:  "Build and run the entire Docker environment",
	Run:    create,
	PreRun: createChecks,
}

func createChecks(cmd *cobra.Command, args []string) {
	network, _ := cmd.Flags().GetString("network")
	datadir, _ := cmd.Flags().GetString("datadir")

	// check flags
	if !isNetworkOk(network) {
		log.WithField("network_flag", network).Fatal("Invalid network")
	}

	if !isDatadirOk(datadir) {
		log.WithField("datadir_flag", datadir).Fatal("Invalid datadir, it must be an absolute path")
	}

	// scratch datadir if not exists
	if err := os.MkdirAll(datadir, 0755); err != nil {
		log.WithError(err).Fatal("An error occured while scratching config dir")
	}

	// check if config file already exists in datadir
	filedir := filepath.Join(datadir, "nigiri.config.json")
	if _, err := os.Stat(filedir); !os.IsNotExist(err) {
		log.WithField("datadir", datadir).Fatal("Configuration file already exists, please delete it first")
	}

	// write and read config file to have viper updated
	if err := config.WriteConfig(filedir); err != nil {
		log.WithError(err).Fatal("An error occured while writing config file")
	}

	if err := config.ReadFromFile(datadir); err != nil {
		log.WithError(err).Fatal("An error occured while reading config file")
	}
}

func create(cmd *cobra.Command, args []string) {
	composePath := getComposePath()

	bashCmd := exec.Command("docker-compose", "-f", composePath, "up", "-d")
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		log.WithError(err).Fatal("An error occured while composing Docker environment")
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
