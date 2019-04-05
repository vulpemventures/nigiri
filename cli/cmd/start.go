package cmd

import (
	"os"
	"os/exec"

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
	composePath := getComposePath()

	bashCmd := exec.Command("docker-compose", "-f", composePath, "start")
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		log.WithError(err).Fatal("An error occured while starting Docker containers")
	}
}

func startStopChecks(cmd *cobra.Command, args []string) {
	datadir, _ := cmd.Flags().GetString("datadir")
	if err := config.ReadFromFile(datadir); err != nil {
		log.Fatal(err)
	}
}
