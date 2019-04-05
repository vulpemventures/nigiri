package cmd

import (
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var StopCmd = &cobra.Command{
	Use:    "stop",
	Short:  "Stop containers",
	Run:    stop,
	PreRun: startStopChecks,
}

func stop(cmd *cobra.Command, args []string) {
	composePath := getComposePath()

	bashCmd := exec.Command("docker-compose", "-f", composePath, "stop")
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		log.WithError(err).Fatal("Error while stopping Docker containers:")
	}
}
