package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vulpemventures/nigiri/cli/config"
)

var DeleteCmd = &cobra.Command{
	Use:    "delete",
	Short:  "Delete all Docker environment components",
	Run:    delete,
	PreRun: deleteChecks,
}

func delete(cmd *cobra.Command, args []string) {
	composePath := getComposePath()

	bashCmd := exec.Command("docker-compose", "-f", composePath, "down")
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		log.WithError(err).Fatal("An error occured while deleting Docker environment")
	}

	if err := cleanVolumes(); err != nil {
		log.WithError(err).Fatal("An error occured while cleanin Docker volumes")
	}

	if err := deleteConfig(); err != nil {
		log.WithError(err).Fatal("An error occured while deleting config file, please delete it manually")
	}
}

func deleteChecks(cmd *cobra.Command, args []string) {
	datadir, _ := cmd.Flags().GetString("datadir")
	if err := config.ReadFromFile(datadir); err != nil {
		log.Fatal(err)
	}
}

/*
	When deleting nigiri we need to clean the Docker volumes that are
	used as datadir of the bitcoin/liquid daemon. These folders contain
	both the *.conf files provided by us and files and directories created
	by the daemons.
	cleanVolumes navigates into <datadir>/resources/volumes/<network>
	and deletes all files and directories but the *.conf config files.
*/
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
