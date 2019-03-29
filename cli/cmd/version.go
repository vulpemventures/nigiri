package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vulpemventures/nigiri/cli/config"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get current tool version",
	Run: func(cmd *cobra.Command, args []string) {
		viper := config.Viper()
		fmt.Println(viper.GetString("version"))
		os.Exit(0)
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		if err := config.ReadFromFile(); err != nil {
			log.Fatal(err)
		}
	},
}
