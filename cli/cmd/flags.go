package cmd

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vulpemventures/nigiri/cli/config"
)

var (
	flagNetwork string
	flagDatadir string
	flagConfig  string
)

var RootCmd = &cobra.Command{
	Use:   "nigiri",
	Short: "Nigiri lets you manage a full dockerized bitcoin environment",
	Long:  "A dockerized environment with a bitcoin and liquid node + block explorer powered by an electrum server",
}

func init() {
	home, _ := homedir.Expand("~")
	defaultDir := filepath.Join(home, ".nigiri")

	RootCmd.PersistentFlags().StringVar(&flagNetwork, "network", "regtest", "Set bitcoin network for containers' services - regtest only for now")
	InitCmd.PersistentFlags().StringVar(&flagDatadir, "datadir", defaultDir, "Set directory for docker containers")

	RootCmd.AddCommand(InitCmd)
	RootCmd.AddCommand(CreateCmd)
	RootCmd.AddCommand(StartCmd)
	RootCmd.AddCommand(StopCmd)
	RootCmd.AddCommand(DeleteCmd)

	viper := config.Viper()
	viper.BindPFlag(config.Network, RootCmd.PersistentFlags().Lookup("network"))
	viper.BindPFlag(config.Datadir, InitCmd.PersistentFlags().Lookup("datadir"))

	cobra.OnInitialize(func() {
		log.SetOutput(os.Stdout)
		log.SetLevel(log.InfoLevel)
	})
}
