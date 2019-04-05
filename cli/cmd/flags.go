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
	flagDatadir      string
	flagNetwork      string
	flagAttachLiquid bool
)

var RootCmd = &cobra.Command{
	Use:   "nigiri",
	Short: "Nigiri lets you manage a full dockerized bitcoin environment",
	Long:  "Nigiri lets you create your dockerized environment with a bitcoin and optionally a liquid node + block explorer powered by an electrum server",
}

func init() {
	defaultDir := getDefaultDir()

	RootCmd.PersistentFlags().StringVar(&flagDatadir, "datadir", defaultDir, "Set directory for config file and docker stuff")
	CreateCmd.PersistentFlags().StringVar(&flagNetwork, "network", "regtest", "Set network for containers' services - regtest only for now")
	CreateCmd.PersistentFlags().BoolVar(&flagAttachLiquid, "liquid", false, "Add liquid sidechain to bitcoin environment")

	RootCmd.AddCommand(CreateCmd)
	RootCmd.AddCommand(StartCmd)
	RootCmd.AddCommand(StopCmd)
	RootCmd.AddCommand(DeleteCmd)
	RootCmd.AddCommand(VersionCmd)

	viper := config.Viper()
	viper.BindPFlag(config.Datadir, RootCmd.PersistentFlags().Lookup("datadir"))
	viper.BindPFlag(config.Network, CreateCmd.PersistentFlags().Lookup("network"))
	viper.BindPFlag(config.AttachLiquid, CreateCmd.PersistentFlags().Lookup("liquid"))

	cobra.OnInitialize(func() {
		log.SetOutput(os.Stdout)
		log.SetLevel(log.InfoLevel)
	})
}

func getDefaultDir() string {
	home, _ := homedir.Expand("~")
	return filepath.Join(home, ".nigiri")
}
