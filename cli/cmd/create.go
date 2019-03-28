package cmd

import (
	"os"

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
	if err := config.ReadFromFile(); err != nil {
		log.Fatal(err)
	}

	if composeExists() {
		log.Fatal("Docker environment already exists, please delete it first")
	}
}

func create(cmd *cobra.Command, args []string) {
	viper := config.Viper()
	network := viper.GetString("network")
	composePath := getComposePath()

	composer := composeBuilder[network](composePath)
	if err := composer.Build(); err != nil {
		log.WithError(err).Fatal("Error while composing Docker environment:")
	}
}

func composeExists() bool {
	_, err := os.Stat(getComposePath())
	return !os.IsNotExist(err)
}
