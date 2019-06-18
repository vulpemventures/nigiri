package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/altafan/nigiri/cli/config"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	flagDatadir       string
	flagNetwork       string
	flagDelete        bool
	flagAttachLiquid  bool
	flagLiquidService bool
	flagEnv           string

	defaultPorts = map[string]map[string]int{
		"bitcoin": {
			"node":        18443,
			"electrs_rpc": 60401,
			"chopsticks":  3000,
			"esplora":     5000,
		},
		"liquid": {
			"node":        7041,
			"electrs_rpc": 51401,
			"chopsticks":  3001,
			"esplora":     5001,
		},
	}
)

var RootCmd = &cobra.Command{
	Use:     "nigiri",
	Short:   "Nigiri lets you manage a full dockerized bitcoin environment",
	Long:    "Nigiri lets you create your dockerized environment with a bitcoin and optionally a liquid node + block explorer powered by an electrum server for every network",
	Version: "0.0.3",
}

func init() {
	defaultDir := getDefaultDir()
	ports, _ := json.Marshal(defaultPorts)

	RootCmd.PersistentFlags().StringVar(&flagDatadir, "datadir", defaultDir, "Set nigiri default directory")
	StartCmd.PersistentFlags().StringVar(&flagNetwork, "network", "regtest", "Set bitcoin network - regtest only for now")
	StartCmd.PersistentFlags().BoolVar(&flagAttachLiquid, "liquid", false, "Enable liquid sidechain")
	StartCmd.PersistentFlags().StringVar(&flagEnv, "ports", string(ports), "Set services ports in JSON format")
	StopCmd.PersistentFlags().BoolVar(&flagDelete, "delete", false, "Stop and delete nigiri")
	LogsCmd.PersistentFlags().BoolVar(&flagLiquidService, "liquid", false, "Set to see logs of a liquid service")

	RootCmd.AddCommand(StartCmd)
	RootCmd.AddCommand(StopCmd)
	RootCmd.AddCommand(LogsCmd)

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
