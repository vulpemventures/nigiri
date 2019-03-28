package cmd

import (
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
	viper := config.Viper()
	network := viper.GetString("network")
	composePath := getComposePath()

	composer := composeBuilder[network](composePath)
	if err := composer.Delete(); err != nil {
		log.WithError(err).Fatal("Error while deleting Docker environment:")
	}
}

func deleteChecks(cmd *cobra.Command, args []string) {
	if err := config.ReadFromFile(); err != nil {
		log.Fatal(err)
	}

	if !composeExists() {
		log.Fatal("Docker environment does not exist")
	}
}
