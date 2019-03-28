package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vulpemventures/nigiri/cli/config"
)

var StartCmd = &cobra.Command{
	Use:    "start",
	Short:  "Start containers",
	Run:    start,
	PreRun: startStopChecks,
}

func start(cmd *cobra.Command, args []string) {
	composePath := filepath.Join(getComposePath(), "docker-compose.yml")
	cmdStart := exec.Command("docker-compose", "-f", composePath, "start")
	cmdStart.Stdout = os.Stdout
	cmdStart.Stderr = os.Stderr

	if err := cmdStart.Run(); err != nil {
		log.WithError(err).Fatal("Error while starting Docker containers:")
	}
}

func startStopChecks(cmd *cobra.Command, args []string) {
	if err := config.ReadFromFile(); err != nil {
		log.Fatal(err)
	}

	if !composeExists() {
		log.Fatal("Docker environment does not exist")
	}
}

func getComposePath() string {
	viper := config.Viper()
	datadir := viper.GetString("datadir")
	network := viper.GetString("network")

	return filepath.Join(datadir, fmt.Sprintf("resources-%s", network))
}
