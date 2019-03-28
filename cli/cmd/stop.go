package cmd

import (
	"os"
	"os/exec"
	"path/filepath"

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
	composePath := filepath.Join(getComposePath(), "docker-compose.yml")
	cmdStart := exec.Command("docker-compose", "-f", composePath, "stop")
	cmdStart.Stdout = os.Stdout
	cmdStart.Stderr = os.Stderr

	if err := cmdStart.Run(); err != nil {
		log.WithError(err).Fatal("Error while stopping Docker containers:")
	}
}
