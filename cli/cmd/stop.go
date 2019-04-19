package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vulpemventures/nigiri/cli/config"
)

var StopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stop and/or delete nigiri",
	RunE:    stop,
	PreRunE: stopChecks,
}

func stopChecks(cmd *cobra.Command, args []string) error {
	datadir, _ := cmd.Flags().GetString("datadir")

	if !isDatadirOk(datadir) {
		return fmt.Errorf("Invalid datadir, it must be an absolute path: %s", datadir)
	}

	if _, err := os.Stat(datadir); os.IsNotExist(err) {
		return fmt.Errorf("Datadir do not exists: %s", datadir)
	}

	nigiriExists, err := nigiriExistsAndNotRunning()
	if err != nil {
		return err
	}
	if !nigiriExists {
		return fmt.Errorf("Nigiri is neither running nor stopped, please create it first")
	}

	if err := config.ReadFromFile(datadir); err != nil {
		return err
	}
	return nil
}

func stop(cmd *cobra.Command, args []string) error {
	delete, _ := cmd.Flags().GetBool("delete")

	bashCmd := getStopBashCmd(delete)
	if err := bashCmd.Run(); err != nil {
		return err
	}

	if delete {
		fmt.Println("Removing volume(s) data")
		if err := cleanVolumes(); err != nil {
			return err
		}
		fmt.Println("Removing configuration file")
		if err := deleteConfig(); err != nil {
			return err
		}
	}

	return nil
}

// cleanVolumes navigates into <datadir>/resources/volumes/<network>
// and deletes all files and directories but the *.conf config files.
func cleanVolumes() error {
	datadir := config.GetString(config.Datadir)
	network := config.GetString(config.Network)
	attachLiquid := config.GetBool(config.AttachLiquid)
	if attachLiquid {
		network = fmt.Sprintf("liquid%s", network)
	}
	volumedir := filepath.Join(datadir, "resources", "volumes", network)

	subdirs, err := ioutil.ReadDir(volumedir)
	if err != nil {
		return err
	}

	for _, d := range subdirs {
		volumedir := filepath.Join(volumedir, d.Name())
		subsubdirs, _ := ioutil.ReadDir(volumedir)
		for _, sd := range subsubdirs {
			if sd.IsDir() {
				if err := os.RemoveAll(filepath.Join(volumedir, sd.Name())); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func deleteConfig() error {
	datadir := config.GetString(config.Datadir)
	configFile := filepath.Join(datadir, config.Filename)
	return os.Remove(configFile)
}

func getStopBashCmd(delete bool) *exec.Cmd {
	composePath := getComposePath()
	bashCmd := exec.Command("docker-compose", "-f", composePath, "stop")
	if delete {
		bashCmd = exec.Command("docker-compose", "-f", composePath, "down")
	}
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	return bashCmd
}
