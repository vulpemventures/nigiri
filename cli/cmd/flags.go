package cmd

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/altafan/nigiri/cli/config"
)

var (
	flagDatadir      string
	flagNetwork      string
	flagDelete       bool
	flagAttachLiquid bool
)

var RootCmd = &cobra.Command{
	Use:     "nigiri",
	Short:   "Nigiri lets you manage a full dockerized bitcoin environment",
	Long:    "Nigiri lets you create your dockerized environment with a bitcoin and optionally a liquid node + block explorer powered by an electrum server for every network",
	Version: "0.1.0",
}

func init() {
	defaultDir := getDefaultDir()

	RootCmd.PersistentFlags().StringVar(&flagDatadir, "datadir", defaultDir, "Set nigiri default directory")
	StartCmd.PersistentFlags().StringVar(&flagNetwork, "network", "regtest", "Set bitcoin network - regtest only for now")
	StartCmd.PersistentFlags().BoolVar(&flagAttachLiquid, "liquid", false, "Enable liquid sidechain")
	StopCmd.PersistentFlags().BoolVar(&flagDelete, "delete", false, "Stop and delete nigiri")

	RootCmd.AddCommand(StartCmd)
	RootCmd.AddCommand(StopCmd)

	viper := config.Viper()
	viper.BindPFlag(config.Datadir, RootCmd.PersistentFlags().Lookup("datadir"))
	viper.BindPFlag(config.Network, StartCmd.PersistentFlags().Lookup("network"))
	viper.BindPFlag(config.AttachLiquid, StartCmd.PersistentFlags().Lookup("liquid"))

	cobra.OnInitialize(func() {
		log.SetOutput(os.Stdout)
		log.SetLevel(log.InfoLevel)
	})
}

func getDefaultDir() string {
	home, _ := homedir.Expand("~")
	return filepath.Join(home, ".nigiri")
}
