package cmd

import (
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vulpemventures/nigiri/cli/config"
	"github.com/vulpemventures/nigiri/cli/constants"
)

var (
	flagDatadir       string
	flagNetwork       string
	flagDelete        bool
	flagAttachLiquid  bool
	flagLiquidService bool
	flagEnv           string
)

var RootCmd = &cobra.Command{
	Use:     "nigiri",
	Short:   "Nigiri lets you manage a full dockerized bitcoin environment",
	Long:    "Nigiri lets you create your dockerized environment with a bitcoin and optionally a liquid node + block explorer powered by an electrum server for every network",
	Version: "0.0.5",
}

func init() {
	c := &config.Config{}
	viper := c.Viper()
	defaultDir := c.GetPath()
	defaultJSON, _ := json.Marshal(constants.DefaultEnv)

	RootCmd.PersistentFlags().StringVar(&flagDatadir, "datadir", defaultDir, "Set nigiri default directory")
	StartCmd.PersistentFlags().StringVar(&flagNetwork, "network", "regtest", "Set bitcoin network - regtest only for now")
	StartCmd.PersistentFlags().BoolVar(&flagAttachLiquid, "liquid", false, "Enable liquid sidechain")
	StartCmd.PersistentFlags().StringVar(&flagEnv, "env", string(defaultJSON), "Set compose env in JSON format")
	StopCmd.PersistentFlags().BoolVar(&flagDelete, "delete", false, "Stop and delete nigiri")

	RootCmd.AddCommand(StartCmd)
	RootCmd.AddCommand(StopCmd)
	RootCmd.AddCommand(LogsCmd)

	LogsCmd.AddCommand(NodeCmd)
	LogsCmd.AddCommand(ElectrsCmd)
	LogsCmd.AddCommand(ChopsticksCmd)
	LogsCmd.AddCommand(EsploraCmd)

	NodeCmd.PersistentFlags().BoolVar(&flagLiquidService, "liquid", false, "Set to see logs of a liquid service")
	ElectrsCmd.PersistentFlags().BoolVar(&flagLiquidService, "liquid", false, "Set to see logs of a liquid service")
	ChopsticksCmd.PersistentFlags().BoolVar(&flagLiquidService, "liquid", false, "Set to see logs of a liquid service")
	EsploraCmd.PersistentFlags().BoolVar(&flagLiquidService, "liquid", false, "Set to see logs of a liquid service")

	viper.BindPFlag(constants.Datadir, RootCmd.PersistentFlags().Lookup("datadir"))
	viper.BindPFlag(constants.Network, StartCmd.PersistentFlags().Lookup("network"))
	viper.BindPFlag(constants.AttachLiquid, StartCmd.PersistentFlags().Lookup("liquid"))

	cobra.OnInitialize(func() {
		log.SetOutput(os.Stdout)
		log.SetLevel(log.InfoLevel)
	})
}
